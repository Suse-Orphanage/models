package models

import (
	"database/sql"

	"gorm.io/gorm"
)

type ThreadLike struct {
	CreateAt sql.NullTime
	ThreadID uint `gorm:"not null;primaryKey"`
	Thread   Thread
	UserID   uint `gorm:"not null;primaryKey"`
	User     User
}

func CreateThreadLike(tid uint, uid uint) error {
	if threadLikedByUser(tid, uid) {
		return NewRequestError("已经点过赞了")
	}

	tl := ThreadLike{
		ThreadID: tid,
		UserID:   uid,
	}
	err := db.Save(tl).Error
	if err != nil {
		return err
	}
	PushThreadLikeNotification(tid, uid)

	err = db.Model(&Thread{}).Update("like_count", gorm.Expr("like_count + ?", 1)).Where("id = ?", tid).Error
	return err
}

func DeleteThreadLike(t *ThreadLike) error {
	if !threadLikedByUser(t.ThreadID, t.UserID) {
		return NewRequestError("没有点过赞")
	}

	err := db.Delete(t).Error
	if err != nil {
		return err
	}
	_ = DeleteNotification(NotificationTypeThreadLike, t.UserID, t.ThreadID)
	err = db.Model(&Thread{}).Update("like_count", gorm.Expr("like_count - ?", 1)).Where("id = ?", t.ThreadID).Error
	return err
}

func DeleteThreadLikeOfThreadForUser(threadID uint, uid uint) error {
	if !threadLikedByUser(threadID, uid) {
		return NewRequestError("没有点过赞")
	}
	err := db.
		Where("thread_id = ? AND user_id = ?", threadID, uid).
		Delete(ThreadLike{}).
		Error
	if err != nil {
		return err
	}
	_ = DeleteNotification(NotificationTypeThreadLike, threadID, uid)

	err = db.Model(&Thread{}).Update("like_count", gorm.Expr("like_count - ?", 1)).Where("id = ?", threadID).Error
	return err
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
