package models

import "time"

type CheckIn struct {
	Year      uint      `gorm:"primaryKey"`
	Month     uint      `gorm:"primaryKey"`
	Day       uint      `gorm:"primaryKey"`
	User      uint      `gorm:"primaryKey" json:"-"`
	ExactTime time.Time `json:"-"`
}

func NewCheckIn(user uint) error {
	now := time.Now()
	record := CheckIn{
		Year:      uint(now.Year()),
		Month:     uint(now.Month()),
		Day:       uint(now.Day()),
		ExactTime: now,
		User:      user,
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
