package models

import (
	"time"

	"gorm.io/gorm"
)

type DynamicConfiguration struct {
	gorm.Model
	Name         string `gorm:"uniqueIndex"`
	Value        string
	LastModified time.Time
}

func GetConfigurationValue(name string) interface{} {
	config := DynamicConfiguration{}
	err := db.Find(&config, "name = ?", name)
	if err != nil {
		return nil
	}
	return config.Value
}
