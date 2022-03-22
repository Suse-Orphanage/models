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

type Session struct {
	gorm.Model
	User   User `json:"-"`
	UserID uint `json:"user_id"`

	Seat   Seat `json:"-"`
	SeatID uint `json:"seat_id"`

	StartTime *time.Time `gorm:"not null"`
	EndTime   *time.Time

	Validate *bool `gorm:"default:true"`

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

func (s *Session) SetValidate(v bool) error {
	s.Validate = &v
	tx := db.Save(s)
	return tx.Error
}

func ValidateSession(seatID uint, start, end *time.Time) bool {
	var cnt int64 = 0
	if end == nil {
		t := start.Add(time.Hour + time.Minute*10)
		end = &t
	}

	// following circumstance is not valid
	// 1. query -> start |-----------| end
	//            |----------|
	//
	// 2. query -> start |-----------| end
	//                        |-----------|
	//
	// 3. query -> start |-----------| end
	//             |-----------------------|
	//
	// 4. query -> start |-----------|end
	//                      |----|

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
		logrus.WithError(tx.Error).Error("Failed to validate session time")
	}
	return cnt == 0 && tx.Error == nil
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
