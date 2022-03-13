package models

import "gorm.io/gorm"

type File struct {
	gorm.Model
	UUID     string `gorm:"type:uuid"`
	Filename string
	Ext      string
}

func NewFile(uuid, filename, ext string) {
	f := File{
		UUID:     uuid,
		Filename: filename,
		Ext:      ext,
	}

	db.Create(&f)
}
