package models

import (
	"time"

	"github.com/sirupsen/logrus"
)

type CheckIn struct {
	Year      uint `gorm:"primaryKey"`
	Month     uint `gorm:"primaryKey"`
	Day       uint `gorm:"primaryKey"`
	UserID    uint `gorm:"primaryKey" json:"-"`
	User      User
	ExactTime time.Time `json:"-"`
}

func NewCheckIn(user uint) error {
	now := time.Now()
	record := CheckIn{
		Year:      uint(now.Year()),
		Month:     uint(now.Month()),
		Day:       uint(now.Day()),
		ExactTime: now,
		UserID:    user,
	}

	var cnt int64
	tx := db.
		Model(&CheckIn{}).
		Where(
			"year = ? AND month = ? AND day = ? AND user_id = ?",
			record.Year,
			record.Month,
			record.Day,
			record.UserID,
		).
		Count(&cnt)

	if cnt != 0 {
		return NewRequestError("今天已经签过到了")
	}

	if tx.Error != nil {
		logrus.WithError(tx.Error).Error("error on handling NewCheckIn")
		return tx.Error
	}

	return db.Create(record).Error
}

func GetCheckInHistory(uid int, beforeYear, beforeMonth, beforeDay int) ([]*CheckIn, error) {
	result := make([]*CheckIn, 0)
	end := time.Date(int(beforeYear), time.Month(beforeMonth), int(beforeDay), 23, 59, 59, 0, time.Local)
	start := end.AddDate(0, 0, -31)
	tx := db.
		Where("user = ? AND ExactTime <= ? AND ExactTime >= ?", uid, end, start).
		Omit("user").
		Order("ExactTime DESC").
		Find(&result)
	return result, tx.Error
}
