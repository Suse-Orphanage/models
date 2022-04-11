package models

import "github.com/google/uuid"

type VisitorMark struct {
	ID   uint   `gorm:"primaryKey"`
	Mark string `gorm:"not null;uniqueIndex"`
}

func CreateVisitorToken() (*VisitorMark, error) {
	uuid := uuid.New()
	token := VisitorMark{
		Mark: uuid.String(),
	}

	err := db.Create(&token).Error
	if err != nil {
		return nil, err
	} else {
		return &token, nil
	}
}

func GetExistingVisitorToken(uuid string) (*VisitorMark, error) {
	result := VisitorMark{}
	err := db.Model(&VisitorMark{}).Where("mark = ?", uuid).First(&result).Error
	return &result, err
}
