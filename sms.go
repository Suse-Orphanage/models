package models

import (
	"math"
	"math/rand"
	"strconv"
	"time"

	"gorm.io/gorm"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type ValidationCodeSms struct {
	gorm.Model
	Code      string `gorm:"type:char(4)"`
	ExpiresAt time.Time
	Phone     string `gorm:"type:char(11);index"`
	Used      bool   `gorm:"type:tinyint;default:false"`
}

func (sms *ValidationCodeSms) SetStatusUsed() {
	sms.Used = true
	db.Save(sms)
}

func FindSmsOfPhone(phone string) *ValidationCodeSms {
	sms := &ValidationCodeSms{}
	r := db.Order("expires_at desc").First(&sms, "phone = ?", phone)

	if r.Error != nil {
		return nil
	}
	return sms
}

func FindSmsOfUser(u *User) *ValidationCodeSms {
	return FindSmsOfPhone(u.Phone)
}

func NewValidationCodeSmsOf(phone string, len uint, expireTime uint) *ValidationCodeSms {
	sms := ValidationCodeSms{
		Phone:     phone,
		Code:      genValidationCode(len),
		ExpiresAt: time.Now().Add(time.Minute * time.Duration(expireTime)),
		Used:      false,
	}
	db.Create(&sms)
	return &sms
}

func NewValidationCodeSmsOfUser(u *User, len, expireTime uint) *ValidationCodeSms {
	return NewValidationCodeSmsOf(u.Phone, len, expireTime)
}

func DeleteValidationCodeSms(sms *ValidationCodeSms) {
	db.Delete(sms, "id = ?", sms.ID)
}

func genValidationCode(len uint) string {
	lowerBound := math.Pow10(int(len - 1))
	upperBound := math.Pow10(int(len))
	code := rand.Intn(int(upperBound)-int(lowerBound)) + int(lowerBound)
	return strconv.Itoa(code)
}
