package models

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	"gorm.io/gorm"
)

type Event struct {
	gorm.Model
	Desc      postgres.Jsonb `gorm:"not null"`
	BeginTime time.Time      `gorm:"type:timestamp;not null"`
	EndTime   time.Time      `gorm:"type:timestamp;not null"`
	Cover     string         `gorm:"type:text;not null"`
	Url       string         `gorm:"type:text;not null"`
}

func GetRecentEvents() []*Event {
	events := make([]*Event, 0)
	now := time.Now()
	begin := now.Add(-time.Hour * 24 * 3)
	db.Where("begin_time > ? AND end_time > ?", begin, now).Order("created_at desc").Find(&events)
	return events
}
