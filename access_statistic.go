package models

import (
	"time"

	"gorm.io/gorm"
)

type AccessStatistic struct {
	gorm.Model
	Time      time.Time
	IP        string `gorm:"type:char(15);index"`
	UserAgent string `gorm:"type:text"`
	Path      string `gorm:"type:text"`
	Method    string `gorm:"type:text"`
	Status    int
	Referer   string `gorm:"type:text"`
	UserID    *uint
	User      *User
}

func AddStatisticInBatch(stats []AccessStatistic) error {
	return db.Create(&stats).Error
}

func GetStatisticInBatchBefore(before uint) ([]AccessStatistic, error) {
	result := make([]AccessStatistic, 0)
	tx := db.
		Model(&AccessStatistic{}).
		Where("id < ?", before).
		Limit(10).
		Order("id desc").
		Find(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return result, nil
}

func GetStatisticInBatchAfter(after uint) ([]AccessStatistic, error) {
	result := make([]AccessStatistic, 0)
	tx := db.
		Model(&AccessStatistic{}).
		Where("id > ?", after).
		Limit(10).
		Order("id asc").
		Find(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return result, nil
}

func GetStatisticInBatchBeforeAfter(before, after uint) ([]AccessStatistic, error) {
	if before >= after {
		return []AccessStatistic{}, nil
	}
	result := make([]AccessStatistic, 0)
	tx := db.
		Model(&AccessStatistic{}).
		Where("id < ? and id > ?", before, after).
		Limit(10).
		Order("id desc").
		Find(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return result, nil
}
