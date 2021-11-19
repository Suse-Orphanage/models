package models

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserType int

func (t *UserType) HasFlag(tp UserType) bool {
	return *t&tp == 1
}

const (
	UserTypePhone UserType = 1 << iota
	UserTypeWx
)

type User struct {
	gorm.Model
	Type        UserType   `gorm:"type:tinyint;notNull"`
	Username    string     `gorm:"type:varchar(32)"`
	Password    string     `gorm:"type:char(64)"`
	Salt        string     `gorm:"type:char(10)"`
	Bio         string     `gorm:"type:varchar(255)"`
	Phone       string     `gorm:"type:char(11);uniqueIndex"`
	Openid      string     `gorm:"type:char(32);index"`
	Unionid     string     `gorm:"type:char(64);index"`
	Session     string     `gorm:"type:varchar(64)"`
	IsPro       bool       `gorm:"notNull;default:0"`
	ProDeadline *time.Time `gorm:"default:NULL"`
	Avatar      string     `gorm:"default:''"`

	RemainingCredit uint `gorm:"default:0"`

	// 被这些人关注
	Followers []*User `gorm:"many2many:user_relations;foreignKey:ID;joinForeignKey:following_id;References:ID;joinReferences:user_id"`
	// 关注了这些人
	Followings []*User `gorm:"many2many:user_relations;foreignKey:ID;joinForeignKey:user_id;References:ID;joinReferences:following_id"`

	LikedThread  []*Thread `gorm:"many2many:user_liked_thread;"`
	StaredThread []*Thread `gorm:"many2many:user_stared_thread;"`
}

type UserRelation struct {
	UserID      int `gorm:"primaryKey"`
	FollowingID int `gorm:"primaryKey"`
	CreatedAt   time.Time
}

func genSalt() string {
	const dict = "AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz1234567890123456789012345678901234567890"
	salt := ""
	for i := 0; i < 10; i++ {
		randIndex := rand.Intn(len(dict))
		salt += string(dict[randIndex])
	}
	return salt
}

func encryptPassword(pass string, salt string) string {
	hash := sha256.New()
	hash.Write([]byte(pass + salt))
	str := hex.EncodeToString(hash.Sum(nil))
	return str
}

func NewWxUser(username, openid, unionid, session string) (*User, error) {
	u := &User{
		Username:    username,
		Type:        UserTypeWx,
		Password:    "",
		Salt:        "",
		Phone:       "",
		Bio:         "",
		Openid:      openid,
		Unionid:     unionid,
		Session:     session,
		IsPro:       false,
		ProDeadline: nil,
		Avatar:      "",
	}
	result := db.Create(&u)
	return u, result.Error
}

func NewPhoneUser(username, phone string) (*User, error) {
	u := &User{
		Username:    username,
		Type:        UserTypePhone,
		Password:    "",
		Salt:        "",
		Phone:       phone,
		Bio:         "",
		Openid:      "",
		Unionid:     "",
		Session:     "",
		IsPro:       false,
		ProDeadline: nil,
		Avatar:      "",
	}
	result := db.Create(&u)
	return u, result.Error
}

func FindWxUser(openid string) (*User, bool) {
	var u User
	result := db.First(&u, "openid = ?", openid)
	return &u, result.Error == nil
}

func FindUser(id uint) (*User, bool) {
	var u User
	result := db.First(&u, "id = ?", id)
	return &u, result.Error == nil
}

func FindUserByPhone(phone string) (*User, bool) {
	var u User
	result := db.First(&u, "phone = ?", phone)
	return &u, result.Error == nil
}

func UpdateUser(id uint, updateField map[string]interface{}) error {
	_, e := FindUser(id)
	if !e {
		return errors.New("用户不存在")
	}

	err := db.
		Model(&User{}).
		Where("id = ?", id).
		Omit("Password", "Salt", "Openid", "Phone", "Unionid", "IsPro", "proDeadline", "RemainingCredit").
		Updates(updateField).
		Error
	if err != nil {
		logrus.WithField("error", err).Panic("Error on querying user.")
		return errors.New("查找用户时出现错误")
	}
	return nil
}

func SetPassword(user *User, password, oldPassword string) error {
	if user.Password != "" {
		validationPassword := encryptPassword(oldPassword, user.Salt)
		if validationPassword != user.Password {
			return NewRequestError("密码错误")
		}
	}

	salt := genSalt()
	tx := db.Model(user).Updates(map[string]interface{}{
		"salt":     salt,
		"password": encryptPassword(password, salt),
	})

	if tx.Error != nil {
		return errors.New("更新密码时出现错误")
	}

	return nil
}

func (u *User) GetAuthBaseInfomation(signup bool) map[string]interface{} {
	return map[string]interface{}{
		"signup":           signup,
		"openid":           u.Openid,
		"userid":           u.ID,
		"bio":              u.Bio,
		"username":         u.Username,
		"avatar":           u.Avatar,
		"is_pro":           u.IsPro,
		"pro_deadline":     u.ProDeadline,
		"remaining_credit": u.RemainingCredit,
		"followers_count":  len(u.Followers),
		"followigns_count": len(u.Followings),
	}
}

func (u *User) GetPhoneAuthBaseInfomation(signup bool) map[string]interface{} {
	return map[string]interface{}{
		"signup":           signup,
		"phone":            u.Phone,
		"userid":           u.ID,
		"bio":              u.Bio,
		"username":         u.Username,
		"avatar":           u.Avatar,
		"is_pro":           u.IsPro,
		"pro_deadline":     u.ProDeadline,
		"remaining_credit": u.RemainingCredit,
		"followers_count":  len(u.Followers),
		"followigns_count": len(u.Followings),
	}
}

func (u *User) GetDetailedInfomation() map[string]interface{} {
	return map[string]interface{}{
		"userid":           u.ID,
		"phone":            u.Phone,
		"avatar":           u.Avatar,
		"username":         u.Username,
		"bio":              u.Bio,
		"is_pro":           u.IsPro,
		"pro_deadline":     u.ProDeadline,
		"remaining_credit": u.RemainingCredit,
		"followers_count":  len(u.Followers),
		"followings_count": len(u.Followers),
	}
}

func (u *User) GetPublicInfomation() map[string]interface{} {
	return map[string]interface{}{
		"userid":           u.ID,
		"avatar":           u.Avatar,
		"username":         u.Username,
		"bio":              u.Bio,
		"is_pro":           u.IsPro,
		"followers_count":  len(u.Followers),
		"followings_count": len(u.Followers),
	}
}

func FollowUser(user, userToBeFollowed *User) error {
	tx := db.Find(&UserRelation{}, "user_id = ? AND following_id = ?", user.ID, userToBeFollowed.ID)
	if tx.Error != nil {
		return errors.New("查询用户时出现错误")
	}
	if tx.RowsAffected != 0 {
		// already followed
		return nil
	}

	user.Followings = append(user.Followings, userToBeFollowed)
	tx = db.Save(user)
	if tx.Error != nil {
		logrus.WithError(tx.Error).Errorf("error on updateing at follow user method.")
		return errors.New("更新用户时出现错误")
	}
	return nil
}

func UnfollowUser(user, userToBeFollowed *User) error {
	r := &UserRelation{}
	tx := db.Find(r, "user_id = ? AND following_id = ?", user.ID, userToBeFollowed.ID)
	if tx.Error != nil {
		return errors.New("查询用户时出现错误")
	}
	if tx.RowsAffected != 1 {
		// already unfollowed
		return nil
	}

	tx = db.Delete(r)
	if tx.Error != nil {
		logrus.WithError(tx.Error).Errorf("error on updating at unfollow user method.")
		return errors.New("更新用户时出现错误")
	}
	return nil
}
