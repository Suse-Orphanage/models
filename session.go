package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"io"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SessionStatus string

const (
	SessionStatusValid    SessionStatus = "valid"
	SessionStatusCanceled SessionStatus = "valid"
	SessionStatusOnGoing  SessionStatus = "on_going"
	SessionStatusExpired  SessionStatus = "expired"
	SessionStatusDone     SessionStatus = "done"
)

type Session struct {
	gorm.Model
	User   User `json:"-"`
	UserID uint `json:"user_id"`

	Seat   Seat `json:"-"`
	SeatID uint `json:"seat_id"`

	StartTime *time.Time `gorm:"not null"`
	EndTime   *time.Time

	ActualEndTime *time.Time
	BillingFee    Price

	Status SessionStatus `gorm:"default:'valid'"`
	// Validate *bool `gorm:"default:true"`

	Token string `gorm:"uniqueIndex,type:varchar(1024)"`
}

func CreateSession(key []byte, u *User, s *Seat, startTime, endTime *time.Time) string {
	uid := u.ID
	uidBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(uidBytes, uint64(uid))
	sid := s.ID
	sidBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(sidBytes, uint64(sid))

	now := time.Now()

	timestamp := make([]byte, 64)
	nanosecond := make([]byte, 64)
	binary.LittleEndian.PutUint64(timestamp, uint64(now.Unix()))
	binary.LittleEndian.PutUint64(nanosecond, uint64(now.Nanosecond()))

	hasher := sha256.New()
	hasher.Write(uidBytes)
	hasher.Write(sidBytes)
	hasher.Write(timestamp)
	hasher.Write(nanosecond)

	hash := hasher.Sum(nil)

	encrypt, err := aes.NewCipher(key)
	if err != nil {
		return ""
	}

	ciphertext := make([]byte, aes.BlockSize+len(hash))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		logrus.WithError(err).Error("failed to generate cipher when creating session")
		return ""
	}

	stream := cipher.NewCFBEncrypter(encrypt, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], hash)

	token := base64.URLEncoding.EncodeToString(ciphertext)

	err = SaveSession(token, u, s, startTime, endTime)
	if err != nil {
		logrus.WithError(err).Error("Failed to save session into database")
		return ""
	}
	return token
}

func SaveSession(token string, u *User, s *Seat, startTime, endTime *time.Time) error {
	tx := db.Create(&Session{
		User:      *u,
		UserID:    u.ID,
		Seat:      *s,
		SeatID:    s.ID,
		Token:     token,
		StartTime: startTime,
		EndTime:   endTime,
	})
	return tx.Error
}

func GetSession(session string) *Session {
	var s Session
	tx := db.Preload("User").Preload("Seat").Where("token = ?", session).First(&s)
	if tx.Error != nil {
		return nil
	}
	return &s
}

func (s *Session) SetEndTime(t *time.Time) error {
	s.EndTime = t
	tx := db.Save(s)
	return tx.Error
}

func (s *Session) SetStatus(status SessionStatus) error {
	s.Status = status
	return db.Save(s).Error
}

// func (s *Session) SetValidate(v bool) error {
// 	s.Validate = &v
// 	tx := db.Save(s)
// 	return tx.Error
// }

func ValidateSession(uid, seatID uint, start, end *time.Time) error {
	var cnt int64 = 0
	if end == nil {
		t := start.Add(time.Hour + time.Minute*10)
		end = &t
	}

	// following circumstance is not valid
	// 1. application -> start |-----------| end
	//                  |----------|
	//
	// 2. application -> start |-----------| end
	//                               |-----------|
	//
	// 3. application-> start |-----------| end
	//                  |-----------------------|
	//
	// 4. application -> start |-----------|end
	//                            |----|
	//
	// 5. user also cannot apply if there exists
	// a session that has been appointed with
	// time range which overlaps with given time range

	tx := db.
		Model(&Session{}).
		Where("seat_id = ?", seatID).
		Where(
			`(start_time <= ? AND end_time <= ? AND end_time >= ?)
			 OR
			 (start_time >= ? AND end_time >= ? AND start_time <= ?)
			 OR
			 (start_time <= ? AND end_time >= ?)
			 OR
			 (start_time >= ? AND end_time <= ?)`,
			start, end, start,
			start, end, end,
			start, end,
			start, end,
		).
		Count(&cnt)

	if tx.Error != nil {
		logrus.WithError(tx.Error).Error("Failed to validate session time when query seat vacancy")
		return tx.Error
	}
	if cnt != 0 {
		return NewRequestError("座位已被预约")
	}

	tx = db.
		Model(&Session{}).
		Where("user_id = ?", uid).
		Where("status = ?", SessionStatusValid).
		Where(
			`(start_time <= ? AND end_time <= ? AND end_time >= ?)
			OR
			(start_time >= ? AND end_time >= ? AND start_time <= ?)
			OR
			(start_time <= ? AND end_time >= ?)
			OR
			(start_time >= ? AND end_time <= ?)`,
			start, end, start,
			start, end, end,
			start, end,
			start, end,
		).
		Count(&cnt)

	if tx.Error != nil {
		logrus.WithError(tx.Error).Error("Failed to validate session time when querying user exising session")
		return tx.Error
	}
	if cnt != 0 {
		return NewRequestError("用户已有预约")
	}

	return nil
}

func GetSessionViaUser(u *User) *Session {
	session := u.Session
	if session == "" {
		return nil
	}

	s := GetSession(session)
	if s == nil {
		return nil
	}
	return s
}

func GetUserSessionHistory(u *User, page int) []*Session {
	sessions := make([]*Session, 0)
	_ = db.Model(&Session{}).
		Where("user_id = ?", u.ID).
		Where("status = ?", SessionStatusValid).
		Where("end_time < ?", time.Now()).
		Update("status", SessionStatusExpired)
	tx := db.
		Find(&sessions, "user_id = ?", u.ID).
		Order("start_time desc").
		Offset(page * 10).
		Limit(10)
	if tx == nil {
		logrus.WithError(tx.Error).Error("Failed to get user session history")
	}
	return sessions
}

func (s *Session) SetBillingFee(fee Price) error {
	s.BillingFee = fee
	return db.Save(s).Error
}

func (s *Session) SetActualEndTime(time *time.Time) error {
	s.ActualEndTime = time
	return db.Save(s).Error
}
