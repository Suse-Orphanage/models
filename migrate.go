package models

import (
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Migrate(connStr string) error {
	logrus.Info("Start migration.")
	db, err := gorm.Open(mysql.Open(connStr), &gorm.Config{})
	if err != nil {
		return err
	}

	err = db.SetupJoinTable(&User{}, "Followings", &UserRelation{})
	if err != nil {
		return err
	}
	err = db.SetupJoinTable(&User{}, "Followers", &UserRelation{})
	if err != nil {
		return err
	}

	return db.AutoMigrate(
		&User{},
		&ValidationCodeSms{},
		&Store{},
		&StoreStar{},
		&Seat{},
		&SeatStatus{},
		&SubscriptionPlan{},
		&Subscription{},
		&Thread{},
		&Coupon{},
		&File{},
		&DynamicConfiguration{},
	)
}
