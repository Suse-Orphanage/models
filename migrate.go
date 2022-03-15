package models

import (
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Migrate(connStr string) error {
	logrus.Info("Start migration.")
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
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
		&DoorNonce{},
		&ThreadStar{},
		&ThreadLike{},
		&Event{},
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
