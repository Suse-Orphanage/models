package models

import (
	"time"

	"gorm.io/gorm"
)

type AccessStatistic struct {
	gorm.Model
	Time      time.Time
	IP        string `gorm:"type:varchar(15);index"`
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

func GetLatestStatistic() ([]AccessStatistic, error) {
	result := make([]AccessStatistic, 0)
	tx := db.
		Model(&AccessStatistic{}).
		Limit(10).
		Order("id desc").
		Find(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return result, nil
}

type ASReport struct {
	Day     time.Time `gorm:"column:day"`
	Count   int       `gorm:"column:cnt"`
	IdStart uint      `gorm:"column:id_start"`
	IdEnd   uint      `gorm:"column:id_end"`
}

func GetHourlyReport() ([]ASReport, error) {
	res := make([]ASReport, 0)
	tx := db.
		Model(&AccessStatistic{}).
		Select("date_trunc('day', time) day, COUNT(*) cnt").
		Group("day").
		Order("day").
		Limit(24).
		Find(&res)
	return res, tx.Error
}

func GetDailyReport() ([]ASReport, error) {
	res := make([]ASReport, 0)
	tx := db.
		Model(&AccessStatistic{}).
		Select("date_trunc('day', time) day, COUNT(*) cnt, MAX(id) id_start, MIN(id) id_end").
		Group("day").
		Order("day").
		Limit(30).
		Find(&res)
	return res, tx.Error
}

func GetMonthlyReport() ([]ASReport, error) {
	res := make([]ASReport, 0)
	tx := db.
		Model(&AccessStatistic{}).
		Select("date_trunc('month', time) month, COUNT(*) cnt").
		Group("month").
		Order("month").
		Limit(3).
		Find(&res)
	return res, tx.Error
}
