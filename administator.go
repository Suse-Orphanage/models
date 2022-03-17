package models

import (
	"gorm.io/gorm"
)

type Administrator struct {
	gorm.Model
	Username string `gorm:"type:text;not null;unique"`
	Password string `gorm:"type:text;not null"`
	Salt     string `gorm:"text;not null"`
	Email    string `gorm:"type:text;not null;unique"`
}

func CreateAdmin(username, password, email string) error {
	salt := genSalt()
	password = encryptPassword(password, salt)
	admin := Administrator{
		Username: username,
		Password: password,
		Salt:     salt,
		Email:    email,
	}
	return db.Create(&admin).Error
}

func (u *Administrator) SetPassword(password string) {
	u.Salt = genSalt()
	u.Password = encryptPassword(password, u.Salt)
}

func (u *Administrator) CheckPassword(password string) bool {
	return u.Password == encryptPassword(password, u.Salt)
}

func MatchAny(username, password string) (*Administrator, error) {
	u := Administrator{}
	tx := db.
		Where("username = ?", username).
		First(&u)
	if tx.Error != nil {
		return nil, NewRequestError("用户不存在")
	}

	if !u.CheckPassword(password) {
		return nil, NewRequestError("密码错误")
	}
	return &u, nil
}
