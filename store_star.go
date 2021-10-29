package models

import "github.com/sirupsen/logrus"

type StoreStar struct {
	StoreID uint
	Store   Store
	UserID  uint
	User    User
}

func StarStore(s Store, u *User) *StoreStar {
	ss := &StoreStar{
		Store: s,
		User:  *u,
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
	list := make([]StoreStar, 10)
	db.Where("user_id = ?", u.ID).Limit(10).Offset((page - 1) * 10).Find(&list)

	stores := make([]Store, 10)
	for _, ss := range list {
		stores = append(stores, ss.Store)
	}
	return stores
}
