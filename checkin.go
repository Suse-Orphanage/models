package models

import "time"

type CheckIn struct {
	Year  uint `gorm:"primaryKey"`
	Month uint `gorm:"primaryKey"`
	Day   uint `gorm:"primaryKey"`
	User  uint `gorm:"primaryKey"`
}

func NewCheckIn(user uint) error {
	record := CheckIn{
		Year:  uint(time.Now().Year()),
		Month: uint(time.Now().Month()),
		Day:   uint(time.Now().Day()),
		User:  user,
	}

	tx := db.First(&CheckIn{},
		"year = ? AND month = ? AND day = ? AND user = ?",
		record.Year,
		record.Month,
		record.Day,
		record.User,
	)

	if tx.Error != nil {
		return NewRequestError("今天已经签过到了")
	}

	return db.Create(record).Error
}
