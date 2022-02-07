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

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type DeviceType uint

const (
	DeviceTypeSeat = iota
	DeviceTypePrinter
	DeviceTypeLamp
	DeviceTypeSocket
	DeviceTypeVendingMachine
)

type DeviceStatus uint

const (
	DeviceStatusUnregistered = iota
	DeviceStatusOffline
	DeviceStatusOnline
	DeviceStatusMaintenance
	DeviceStatusOccupied
)

type Device struct {
	gorm.Model
	DeviceID     string       `gorm:"uniqueIndex;type:varchar(128)"`
	Name         string       `gorm:"type:varchar(128)"`
	Type         DeviceType   `gorm:"type:tinyint;default:0"`
	Status       DeviceStatus `gorm:"type:tinyint;default:0"`
	Seat         Seat
	SeatID       uint
	ConnectionID string `gorm:"type:varchar(128);uniqueIndex"`
	CurrentToken string `gorm:"type:varchar(1024)"`
}

type DeviceToken struct {
	gorm.Model
	Affiliate   Device
	AffiliateID uint
	Token       string
	Valid       bool      `gorm:"defualt:true"`
	Deadline    time.Time `gorm:"not null"`
}

func (d *Device) CreateToken(key []byte, expiration uint) string {
	id := d.DeviceID
	exp := time.Now().Add(time.Duration(expiration) * time.Minute)

	hasher := sha256.New()

	timestamp := make([]byte, 64)
	nanosecond := make([]byte, 64)
	binary.LittleEndian.PutUint64(timestamp, uint64(exp.Unix()))
	binary.LittleEndian.PutUint64(nanosecond, uint64(exp.Nanosecond()))

	hasher.Write([]byte(id))
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

	err = SaveToken(d, token, exp)
	if err != nil {
		logrus.WithError(err).Error("Failed to save device token into database")
		return ""
	}
	return token
}

func SaveToken(device *Device, token string, expiration time.Time) error {
	tx := db.Create(&DeviceToken{
		AffiliateID: device.ID,
		Token:       token,
		Valid:       true,
		Deadline:    expiration,
	})
	return tx.Error
}

func (t *DeviceToken) SetValid(valid bool) error {
	t.Valid = valid
	tx := db.Save(t)
	return tx.Error
}

func (t *DeviceToken) IsExpired() bool {
	return t.Valid && time.Now().After(t.Deadline)
}

func (d *Device) SetDeviceStatus(status DeviceStatus) error {
	d.Status = status
	tx := db.Save(d)
	return tx.Error
}

func (d *Device) SetConnectionID() error {
	d.ConnectionID = uuid.New().String()
	tx := db.Save(d)
	return tx.Error
}

func GetDeviceByID(deviceId string) *Device {
	d := Device{}
	tx := db.First(&d, "device_id = ?", deviceId)
	if tx.Error != nil {
		return nil
	}
	return &d
}

func GetDeviceToken(token string) *DeviceToken {
	t := DeviceToken{}
	tx := db.Preload("Device").First(&t, "token = ?", token)
	if tx.Error != nil {
		return nil
	}
	return &t
}
