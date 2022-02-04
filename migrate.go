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

	err = db.AutoMigrate(
		&User{},
		&ValidationCodeSms{},
		&Store{},
		&StoreStar{},
		&Seat{},
		&SeatStatus{},
		&Good{},
		&Order{},
		&Thread{},
		&CheckIn{},
		&Coupon{},
		&File{},
		&DynamicConfiguration{},
		&Device{},
		&DeviceToken{},
	)

	if err != nil {
		return err
	}

	goodsList := *GetBuiltinGoods()
	for idx, good := range goodsList {
		if db.Find(&Good{}, good.ID).RowsAffected == 0 {
			tx := db.Create(&goodsList[idx])
			if tx.Error != nil {
				return tx.Error
			}
		}
	}

	return nil
}
