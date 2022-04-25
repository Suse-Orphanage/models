package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type NotificationType string

const (
	NotificationTypeThreadReply         = "thread_reply"
	NotificationTypeThreadLike          = "thread_like"
	NotificationTypeFollows             = "follow"
	NotificationTypeBeMentioned         = "be_mentioned"
	NotificationTypeChat                = "chat"
	NotificationTypeCreditsRunOut       = "credits_run_out"
	NotificationTypeScheduledTimeArrive = "scheduled_time_arrive"
	NotificationTypeFollowingOnline     = "following_online"
	NotificationTypeLinkedMessage       = "linked_message"
)

type Notification struct {
	gorm.Model
	Type NotificationType `gorm:"type:string;not null"`
	// used to store affiliate subject,
	// i.e. thread, reply, message etc., speeds up
	// query when deleting the notification.
	AffiliateNotificationSubjectID *uint
	User                           *User
	UserID                         uint
	Data                           json.RawMessage `gorm:"type:jsonb"`
}

func PushNotification(n *Notification) error {
	err := db.Save(n).Error
	return err
}

func DeleteNotifications(t NotificationType, uid, aff_id uint) error {
	return db.Delete(
		"type = ? AND user_id = ? AND affiliate_notification_subject_id = ?",
		t,
		uid,
		aff_id,
	).Error
}

func QueryNotificationOfType(t NotificationType, u User, limit, page int) ([]*Notification, error) {
	result := make([]*Notification, 0)
	err := db.
		Where("type = ? AND user_id = ?", t, u.ID).
		Limit(limit).
		Offset(limit * (page - 1)).
		Order("created_at desc").
		Find(&result).
		Error
	return result, err
}

func QueryNotification(u User, limit, page int) ([]*Notification, error) {
	notifications := make([]*Notification, 0)
	err := db.
		Where("user_id = ? AND created_at > ?", u.ID, u.LatestNotificationReadTime).
		Limit(limit).
		Offset(limit * (page - 1)).
		Order("created_at desc").
		Find(&notifications).
		Error
	return notifications, err
}

func (u *User) CommitNotificationRead(t time.Time) error {
	if t.After(time.Now().Add(time.Second * 5)) {
		return NewRequestError("时间不正确")
	}
	u.LatestNotificationReadTime = t
	return db.Save(u).Error
}