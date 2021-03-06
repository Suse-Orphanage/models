package models

import "github.com/sirupsen/logrus"

type StoreStar struct {
	StoreID uint `gorm:"primaryKey"`
	Store   Store
	UserID  uint `gorm:"primaryKey"`
	User    User
}

func StarStore(s Store, u *User) *StoreStar {
	ss := &StoreStar{
		StoreID: s.ID,
		UserID:  u.ID,
	}
	tx := db.FirstOrCreate(ss)
	if tx.Error != nil {
		logrus.WithError(tx.Error).Error("error on recording stared store")
		return nil
	}
	return ss
}

func UnstarStore(s Store, u *User) bool {
	tx := db.Delete(&StoreStar{}, "store_id = ? AND user_id = ?", s.ID, u.ID)
	if tx.Error != nil {
		logrus.WithError(tx.Error).Error("error on deleting stared store")
		return false
	}
	return true
}

func GetStaredStores(u *User, page int) []Store {
	var list []StoreStar
	tx := db.Preload("Store").Where("user_id = ?", u.ID).Limit(10).Offset((page - 1) * 10).Find(&list)
	if tx.Error != nil {
		logrus.WithError(tx.Error).Error("error when querying for user stared store")
		return make([]Store, 0)
	}

	var stores []Store
	for _, ss := range list {
		stores = append(stores, ss.Store)
	}
	return stores
}
