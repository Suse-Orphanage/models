package models

import "database/sql"

type ThreadStar struct {
	CreateAt sql.NullTime
	ThreadID uint `gorm:"not null;primaryKey"`
	Thread   Thread
	UserID   uint `gorm:"not null;primaryKey"`
	User     User
}

func CreateThreadStar(t *Thread, u *User) error {
	tl := ThreadStar{
		ThreadID: t.ID,
		UserID:   u.ID,
	}

	return db.Save(tl).Error
}

func DeleteThreadStar(t *ThreadStar) error {
	return db.Delete(t).Error
}

func DeleteThreadStarOfThreadForUser(threadID uint, uid uint) error {
	return db.
		Where("thread_id = ? AND user_id = ?", threadID, uid).
		Delete(ThreadStar{}).
		Error
}

func FindThreadStarForUser(uid uint) ([]ThreadStar, error) {
	var tls []ThreadStar
	err := db.Where("user_id = ?", uid).Find(&tls).Error
	return tls, err
}

func FindThreadStarCount(threadID uint) (uint, error) {
	var count int64 = 0
	err := db.Model(ThreadStar{}).Where("thread_id = ?", threadID).Count(&count).Error
	return uint(count), err
}
