package models

import (
	"database/sql"
)

type ThreadLike struct {
	CreateAt sql.NullTime
	ThreadID uint `gorm:"not null;primaryKey"`
	Thread   Thread
	UserID   uint `gorm:"not null;primaryKey"`
	User     User
}

func CreateThreadLike(tid uint, uid uint) error {
	tl := ThreadLike{
		ThreadID: tid,
		UserID:   uid,
	}
	return db.Save(tl).Error
}

func DeleteThreadLike(t *ThreadLike) error {
	return db.Delete(t).Error
}

func DeleteThreadLikeOfThreadForUser(threadID uint, uid uint) error {
	return db.
		Where("thread_id = ? AND user_id = ?", threadID, uid).
		Delete(ThreadLike{}).
		Error
}

func FindThreadLikeForUser(uid uint) ([]ThreadLike, error) {
	var tls []ThreadLike
	err := db.Where("user_id = ?", uid).Find(&tls).Error
	return tls, err
}

func FindThreadLikeCount(threadID uint) uint {
	var count int64 = 0
	_ = db.Model(ThreadLike{}).Where("thread_id = ?", threadID).Count(&count).Error
	return uint(count)
}

func threadLikedByUser(threadId, userId uint) bool {
	var count int64 = 0
	_ = db.Model(ThreadLike{}).Where("thread_id = ? AND user_id = ?", threadId, userId).Count(&count).Error
	return count > 0
}
