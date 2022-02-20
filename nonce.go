package models

import (
	"fmt"
	"math/rand"
	"time"
)

type DoorNonce struct {
	Nonce string `gorm:"index;primaryKey"`

	CreationTime time.Time `json:"creation_time"`
	ExpireTime   time.Time `json:"expire_time"`

	User      User
	UserID    uint `gorm:"not null"`
	Session   *Session
	SessionID *uint

	Valid bool `gorm:"default:true;primaryKey"`
}

func CreateDoorNonce(user *User, session *Session) *DoorNonce {
	if n := GetRecentValidNonce(user); n != nil {
		return n
	}

	work := false
	var d *DoorNonce = nil
	var nonce string

	for !work {
		n := rand.Int() % 10000
		nonce = fmt.Sprintf("%04d", n)

		d = &DoorNonce{
			Nonce:  nonce,
			User:   *user,
			UserID: user.ID,
			Valid:  true,

			CreationTime: time.Now(),
			ExpireTime:   time.Now().Add(time.Hour / 2),
		}
		if session != nil {
			d.SessionID = &session.ID
		} else {
			d.SessionID = nil
		}

		tx := db.Create(d)
		work = tx.Error == nil
	}
	return d
}

func CleanThoseExpired() {
	nonces := []DoorNonce{}
	_ = db.Where(&nonces, "valid = ? AND expire_time > ?", true, time.Now()).Update("valid", false)
}

func GetDoorNonce(nonce string) (*DoorNonce, error) {
	d := &DoorNonce{}
	tx := db.Where("nonce = ? AND valid = ?", nonce, true).First(d)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return d, nil
}

func (n *DoorNonce) SetNoLongerValid() error {
	tx := db.Model(n).Update("valid", false)
	return tx.Error
}

func UserHasValidNonceBefore(u *User) bool {
	CleanThoseExpired()

	nonces := []DoorNonce{}
	_ = db.Find(&nonces, "user_id = ? AND valid = ?", u.ID, true)

	return len(nonces) != 0
}

func GetRecentValidNonce(u *User) *DoorNonce {
	nonce := &DoorNonce{}
	tx := db.First(&nonce, "user_id = ? AND valid = ? AND expire_time > ?", u.ID, true, time.Now())
	if tx.RowsAffected == 0 {
		return nil
	}
	return nonce
}
