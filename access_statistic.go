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
	User      *User
}

func AddStatisticInBatch(stats []AccessStatistic) {
	db.Create(&stats)
}
