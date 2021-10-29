package models

import "gorm.io/gorm"

type Thread struct {
	gorm.Model
	Content   string `gorm:"default:{}"`
	Likes     uint   `gorm:"default:0"`
	Title     string `gorm:"type:varchar(20)"`
	LikeCount uint
	StarCount uint
	ParentID  uint
	Parent    *Thread `gorm:"foreignKey:ParentID;default:null;"`
	AuthorID  uint
	Author    *User `gorm:"foreignKey:AuthorID"`
	Level     int   `gorm:"type:tinyint(1);default:1"`

	LikedUser  []*User `gorm:"many2many:user_liked_thread;"`
	StaredUser []*User `gorm:"many2many:user_stared_thread;"`
}
