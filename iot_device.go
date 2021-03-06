package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"io"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type DeviceKind uint

const (
	DeviceKindSeat = iota
	DeviceKindPrinter
	DeviceKindLamp
	DeviceKindSocket
	DeviceKindVendingMachine
	DeviceKindOther
)

func (t *DeviceKind) MarshalJSON() ([]byte, error) {
	tStr := ""
	switch *t {
	case DeviceKindSeat:
		tStr = "seat"
	case DeviceKindPrinter:
		tStr = "printer"
	case DeviceKindLamp:
		tStr = "lamp"
	case DeviceKindSocket:
		tStr = "socket"
	case DeviceKindVendingMachine:
		tStr = "vending_machine"
	default:
		tStr = "other"
	}
	return []byte(`"` + tStr + `"`), nil
}

type DeviceStatus uint

const (
	DeviceStatusUnregistered = iota
	DeviceStatusOffline
	DeviceStatusOnline
	DeviceStatusMaintenance
	DeviceStatusOccupied
)

func (s *DeviceStatus) MarshalJSON() ([]byte, error) {
	sStr := ""
	switch *s {
	case DeviceStatusUnregistered:
		sStr = "unregistered"
	case DeviceStatusOffline:
		sStr = "offline"
	case DeviceStatusOnline:
		sStr = "online"
	case DeviceStatusMaintenance:
		sStr = "maintenance"
	case DeviceStatusOccupied:
		sStr = "occupied"
	}
	return []byte(`"` + sStr + `"`), nil
}

type Device struct {
	gorm.Model
	DeviceID     string       `gorm:"uniqueIndex;type:varchar(128)"`
	Name         string       `gorm:"type:varchar(128)"`
	Kind         DeviceKind   `gorm:"type:int;default:0"`
	Status       DeviceStatus `gorm:"type:int;default:0"`
	Seat         *Seat
	SeatID       *uint
	ConnectionID *string `gorm:"type:varchar(128);uniqueIndex"`
	CurrentToken *string `gorm:"type:varchar(1024)"`

	LastActiveAt *time.Time `gorm:"type:timestamp"`

	ExpectedStatus *json.RawMessage `gorm:"type:jsonb"`
}

type DeviceToken struct {
	gorm.Model
	Affiliate   Device
	AffiliateID uint
	User        User
	UserID      uint `gorm:"not null"`
	Session     Session
	SessionID   uint `gorm:"not null"`
	Token       string
	Valid       *bool     `gorm:"defualt:true"`
	Deadline    time.Time `gorm:"not null"`
}

func (d *Device) CreateToken(key []byte, expiration uint, u *User, s *Session) string {
	id := d.DeviceID
	exp := time.Now().Add(time.Duration(expiration) * time.Minute)

	hasher := sha256.New()

	timestamp := make([]byte, 64)
	nanosecond := make([]byte, 64)
	binary.LittleEndian.PutUint64(timestamp, uint64(exp.Unix()))
	binary.LittleEndian.PutUint64(nanosecond, uint64(exp.Nanosecond()))

	hasher.Write([]byte(id))
	hasher.Write([]byte(s.Token))
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

	err = SaveToken(d, token, exp, u, s)
	if err != nil {
		logrus.WithError(err).Error("Failed to save device token into database")
		return ""
	}
	return token
}

func SaveToken(device *Device, token string, expiration time.Time, u *User, s *Session) error {
	tx := db.Create(&DeviceToken{
		AffiliateID: device.ID,
		Token:       token,
		Deadline:    expiration,
		UserID:      u.ID,
		SessionID:   s.ID,
	})
	if tx.Error != nil {
		return tx.Error
	}
	device.CurrentToken = &token
	return db.Save(&device).Error
}

func EmptyToken(device *Device) error {
	tx := db.Model(device).Update("current_token", nil)
	return tx.Error
}

func (t *DeviceToken) SetValid(valid bool) error {
	t.Valid = &valid
	tx := db.Save(t)
	return tx.Error
}

func (t *DeviceToken) IsExpired() bool {
	return *t.Valid && time.Now().After(t.Deadline)
}

func (d *Device) SetDeviceStatus(status DeviceStatus) error {
	d.Status = status
	tx := db.Save(d)
	return tx.Error
}

func (d *Device) SetConnectionID(id string) error {
	d.ConnectionID = &id
	tx := db.Save(d)
	return tx.Error
}

func (d *Device) EmptyConnectID() error {
	d.ConnectionID = nil
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
