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

type UserStatus uint

const (
	UserStatusNormal UserStatus = iota
	UserStatusBanned
	UserStatusDeleted
)

func (us *UserStatus) MarshalJSON() ([]byte, error) {
	var status string
	switch *us {
	case UserStatusNormal:
		status = "normal"
	case UserStatusBanned:
		status = "banned"
	case UserStatusDeleted:
		status = "deleted"
	}
	return []byte(`"` + status + `"`), nil
}

type UserBillingStatus uint

const (
	UserBillingStatusNone UserBillingStatus = iota
	UserBillingStatusOnBilling
)

func (us *UserBillingStatus) MarshalJSON() ([]byte, error) {
	var status string
	switch *us {
	case UserBillingStatusOnBilling:
		status = "billing"
	case UserBillingStatusNone:
		status = "none"
	}
	return []byte(`"` + status + `"`), nil
}

type User struct {
	gorm.Model
	Type        UserType   `gorm:"type:int;notNull" json:"-"`
	Username    string     `gorm:"type:varchar(32)" json:"username"`
	Password    string     `gorm:"type:varchar(64)" json:"-"`
	Salt        string     `gorm:"type:varchar(10)" json:"-"`
	Bio         string     `gorm:"type:varchar(255)" json:"bio"`
	Phone       string     `gorm:"type:varchar(11);uniqueIndex" json:"-"`
	Openid      string     `gorm:"type:varchar(32);index" json:"-"`
	Unionid     string     `gorm:"type:varchar(64);index" json:"-"`
	WxSession   string     `gorm:"type:varchar(64)" json:"-"`
	IsPro       bool       `gorm:"notNull;default:false" json:"is_pro"`
	ProDeadline *time.Time `gorm:"default:NULL" json:"-"`
	Avatar      string     `gorm:"default:''" json:"avatar"`

	CurrentOccupiedSeat   *Seat `json:"-"`
	CurrentOccupiedSeatID *uint `json:"-"`

	Status              UserStatus        `gorm:"default:0" json:"-"`
	BillingStatus       UserBillingStatus `gorm:"default:0" json:"-"`
	RecentBillStartTime *time.Time        `gorm:"default:NULL" json:"-"`
	Session             string            `gorm:"type:text;default:''" json:"-"`

	RemainingCredit Price `gorm:"default:0" json:"-"`

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
		WxSession:   session,
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
		WxSession:   "",
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

func (u *User) SetAvatar(avatar string) error {
	tx := db.Model(u).Updates(map[string]interface{}{
		"avatar": avatar,
	})

	if tx.Error != nil {
		return errors.New("更新头像时出现错误")
	}

	return nil
}

func (u *User) CheckPassword(password string) bool {
	return encryptPassword(password, u.Salt) == u.Password
}

func (u *User) GetFollowersCount() int64 {
	var count int64 = 0
	db.
		Table("user_relations").
		Where("following_id = ?", u.ID).
		Count(&count)
	return count
}

func (u *User) GetFollowingsCount() int64 {
	var count int64 = 0
	db.
		Table("user_relations").
		Where("user_id = ?", u.ID).
		Count(&count)
	return count
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
		"remaining_credit": u.RemainingCredit.ToFloat64(),
		"followers_count":  u.GetFollowersCount(),
		"followings_count": u.GetFollowingsCount(),
		"password_set":     u.Password != "",
		"status":           u.Status,
		"billing_status":   u.BillingStatus,
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
		"remaining_credit": u.RemainingCredit.ToFloat64(),
		"followers_count":  u.GetFollowersCount(),
		"followings_count": u.GetFollowingsCount(),
		"password_set":     u.Password != "",
		"status":           u.Status,
		"billing_status":   u.BillingStatus,
	}
}

func (u *User) GetDetailedInfomation(whoInquery *User) map[string]interface{} {
	var tmp int64 = 0
	db.
		Table("user_relations").
		Where("user_id = ? and following_id = ?", whoInquery.ID, u.ID).
		Count(&tmp)
	isFollowing := tmp == 1
	db.
		Table("user_relations").
		Where("user_id = ? and following_id = ?", u.ID, whoInquery.ID).
		Count(&tmp)
	isFollower := tmp == 1

	return map[string]interface{}{
		"userid":           u.ID,
		"phone":            u.Phone,
		"avatar":           u.Avatar,
		"username":         u.Username,
		"bio":              u.Bio,
		"is_pro":           u.IsPro,
		"followers_count":  u.GetFollowersCount(),
		"followings_count": u.GetFollowingsCount(),
		"status":           u.Status,
		"is_following":     isFollowing,
		"is_follower":      isFollower,
		// "pro_deadline":     u.ProDeadline,
		// "remaining_credit": u.RemainingCredit.ToFloat64(),
		// "billing_status":   u.BillingStatus,
	}
}

type UserPublicInfomation struct {
	ID              uint       `json:"userid"`
	Username        string     `json:"username"`
	Avatar          string     `json:"avatar"`
	Bio             string     `json:"bio"`
	IsPro           bool       `json:"is_pro"`
	FollowersCount  int64      `json:"followers_count"`
	FollowingsCount int64      `json:"followings_count"`
	Status          UserStatus `json:"status"`
}

func (u *User) GetPublicInfomation() UserPublicInfomation {
	return UserPublicInfomation{
		ID:              u.ID,
		Username:        u.Username,
		Avatar:          u.Avatar,
		Bio:             u.Bio,
		IsPro:           u.IsPro,
		FollowersCount:  u.GetFollowersCount(),
		FollowingsCount: u.GetFollowingsCount(),
		Status:          u.Status,
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

func (u *User) IncreaseCreditBy(cnt float64) error {
	u.RemainingCredit += Price(uint64(u.RemainingCredit) + uint64(cnt*100))
	tx := db.Save(u)
	if tx.Error != nil {
		logrus.WithError(tx.Error).Errorf("error on updating at increase credit method.")
		return errors.New("更新用户时出现错误")
	}
	return nil
}

func (u *User) DecreaseCreditBy(cnt float64) error {
	remainingCredit := int64(u.RemainingCredit) - int64(cnt*100)
	if remainingCredit < 0 {
		remainingCredit = 0
	}
	u.RemainingCredit = Price(remainingCredit)
	tx := db.Save(u)
	if tx.Error != nil {
		logrus.WithError(tx.Error).Errorf("error on updating at decrease credit method.")
		return errors.New("更新用户时出现错误")
	}
	return nil
}

func (u *User) CreditLessThan(v float64) bool {
	return u.RemainingCredit.ToFloat64() < v
}

func (u *User) SetStatus(s UserStatus) error {
	u.Status = s
	return db.Save(u).Error
}

func (u *User) SetBillingStatus(s UserBillingStatus) error {
	u.BillingStatus = s
	return db.Save(u).Error
}

func (u *User) SetRecentBillTime(t *time.Time) error {
	u.RecentBillStartTime = t
	return db.Save(u).Error
}

func (u *User) SetProDeadline(t *time.Time) error {
	u.ProDeadline = t
	return db.Save(u).Error
}

func (u *User) SetCurrentOccupiedSeat(seat *Seat) error {
	if seat == nil {
		u.CurrentOccupiedSeatID = nil
	} else {
		u.CurrentOccupiedSeatID = &seat.ID
	}

	return db.Save(u).Error
}

func (u *User) GetCurrentOccupiedDevices() []Device {
	ret := make([]Device, 0)
	if u.CurrentOccupiedSeatID != nil {
		return ret
	}

	tx := db.Find(&ret, "seat_id = ?", u.CurrentOccupiedSeatID)
	if tx.Error != nil {
		return []Device{}
	}

	return ret
}

func (u *User) SetSession(s *Session) error {
	u.Session = s.Token
	return db.Save(u).Error
}

func (u *User) ClearSession() error {
	u.Session = ""
	return db.Save(u).Error
}

func (u *User) SetWxSession(s string) error {
	u.WxSession = s
	return db.Save(u).Error
}

type BasicUserInfomation struct {
	ID       uint64 `json:"userid"`
	Avatar   string `json:"avatar"`
	Username string `json:"username"`
	Bio      string `json:"bio"`
	IsPro    bool   `json:"is_pro"`
}

func (u *User) GetFollowers() ([]BasicUserInfomation, error) {
	ret := make([]BasicUserInfomation, 0)
	tx := db.
		Table("user_relations").
		Find(&ret, "following_id = ?", u.ID)
	if tx.Error != nil {
		return []BasicUserInfomation{}, tx.Error
	}

	return ret, nil
}

func (u *User) GetFollowings() ([]BasicUserInfomation, error) {
	ret := make([]BasicUserInfomation, 0)
	tx := db.
		Table("user_relations").
		Find(&ret, "user_id = ?", u.ID)
	if tx.Error != nil {
		return []BasicUserInfomation{}, tx.Error
	}

	return ret, nil
}

func GetUserCount() uint {
	var count int64
	db.Model(&User{}).Count(&count)
	return uint(count)
}
