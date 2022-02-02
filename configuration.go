package models

import (
	"time"

	"gorm.io/gorm"
)

type DynamicConfiguration struct {
	gorm.Model
	Name         string `gorm:"uniqueIndex;type:varchar(128)"`
	Value        string
	LastModified time.Time
}

func GetConfigurationValue(name string) interface{} {
	config := DynamicConfiguration{}
	tx := db.First(&config, "name = ?", name)
	if tx.Error != nil {
		return nil
	}
	return config.Value
}
