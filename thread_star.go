package models

import (
	"database/sql"

	"gorm.io/gorm"
)

type ThreadStar struct {
	CreateAt sql.NullTime
	ThreadID uint `gorm:"not null;primaryKey"`
	Thread   Thread
	UserID   uint `gorm:"not null;primaryKey"`
	User     User
}

func CreateThreadStar(threadId, userId uint) error {
	if threadStaredByUser(threadId, userId) {
		return NewRequestError("已经收藏过了")
	}

	tl := ThreadStar{
		ThreadID: threadId,
		UserID:   userId,
	}

	err := db.Save(tl).Error
	if err != nil {
		return err
	}

	err = db.Model(&Thread{}).Update("star_count", gorm.Expr("star_count + ?", 1)).Where("id = ?", threadId).Error
	return err
}

func DeleteThreadStar(t *ThreadStar) error {
	if !threadStaredByUser(t.ThreadID, t.UserID) {
		return NewRequestError("没有收藏过")
	}

	err := db.Delete(t).Error
	if err != nil {
		return err
	}
	err = db.Model(&Thread{}).Update("star_count", gorm.Expr("star_count - ?", 1)).Where("id = ?", t.ThreadID).Error
	return err
}

func DeleteThreadStarOfThreadForUser(threadID uint, uid uint) error {
	if !threadStaredByUser(threadID, uid) {
		return NewRequestError("没有收藏过")
	}
	err := db.
		Where("thread_id = ? AND user_id = ?", threadID, uid).
		Delete(ThreadStar{}).
		Error
	if err != nil {
		return err
	}

	err = db.Model(&Thread{}).Update("star_count", gorm.Expr("star_count - ?", 1)).Where("id = ?", threadID).Error
	return err
}

func FindThreadStarForUser(uid uint) ([]ThreadStar, error) {
	var tls []ThreadStar
	err := db.Where("user_id = ?", uid).Find(&tls).Error
	return tls, err
}

func FindThreadStarCount(threadID uint) uint {
	var count int64 = 0
	_ = db.Model(ThreadStar{}).Where("thread_id = ?", threadID).Count(&count).Error
	return uint(count)
}

func threadStaredByUser(threadId, userId uint) bool {
	var count int64 = 0
	_ = db.Model(ThreadStar{}).Where("thread_id = ? AND user_id = ?", threadId, userId).Count(&count).Error
	return count > 0
}

func GetUserStaredThreads(uid uint, page int) ([]*Post, error) {
	const perPage = 10
	threadStars := make([]ThreadStar, 0)
	tx := db.Where("user_id = ?", uid).Limit(perPage).Offset(perPage * (page - 1)).Find(&threadStars)

	posts := make([]*Post, len(threadStars))
	for i, star := range threadStars {
		thread := GetThreadByID(star.ThreadID)
		posts[i] = ConstructPostObject(*thread, uid)
	}

	return posts, tx.Error
}
