package models

import (
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
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

func (n *Notification) MarshalJSON() ([]byte, error) {
	r := make(map[string]interface{}, 0)
	r["type"] = n.Type
	r["time"] = n.CreatedAt
	r["data"] = n.Data
	switch n.Type {
	case NotificationTypeThreadReply:
		fallthrough
	case NotificationTypeThreadLike:
		if n.AffiliateNotificationSubjectID == nil {
			logrus.Errorf("Notification.MarshalJSON: notification id %d has no affiliate subject id", n.ID)
			return json.Marshal(nil)
		}
		post := GetPostSummary(*n.AffiliateNotificationSubjectID)
		if post == nil {
			logrus.Errorf("PushThreadReplyNotification: failed to get post summary, post object is nil")
			return json.Marshal(nil)
		}
		r["thread_id"] = post.ID
		r["thread_title"] = post.Title
		r["thread_content"] = post.Content
		r["thread_author"] = post.Author
	case NotificationTypeFollows:
		var u *User
		if n.AffiliateNotificationSubjectID != nil {
			u, _ = FindUser(n.UserID)
			r["follower"] = u.GetPublicInfomation()
		} else {
			logrus.Errorf("Notification.MarshalJSON: notification id %d has no follower id", n.ID)
		}
	case NotificationTypeBeMentioned:
		// TODO: to be implemented
		logrus.Error("PushBeMentionedNotification: not implemented")
	case NotificationTypeChat:
		var u *User
		if n.AffiliateNotificationSubjectID != nil {
			u, _ = FindUser(n.UserID)
			r["sender"] = u.GetPublicInfomation()
		} else {
			logrus.Errorf("Notification.MarshalJSON: notification id %d has no sender id", n.ID)
		}
	case NotificationTypeCreditsRunOut:
		break
	case NotificationTypeScheduledTimeArrive:
		// TODO: add related session
		break
	case NotificationTypeFollowingOnline:
		var u *User
		if n.AffiliateNotificationSubjectID != nil {
			u, _ = FindUser(n.UserID)
			r["following"] = u.GetPublicInfomation()
		} else {
			logrus.Errorf("Notification.MarshalJSON: notification id %d has no friends id", n.ID)
		}
	}
	return json.Marshal(r)
}

func PushNotification(n *Notification) error {
	err := db.Save(n).Error
	return err
}

func DeleteNotification(t NotificationType, uid, aff_id uint) error {
	return db.Delete(
		"type = ? AND user_id = ? AND affiliate_notification_subject_id = ?",
		t,
		uid,
		aff_id,
	).Error
}

func DeleteNotifications(t NotificationType, aff_id uint) error {
	return db.Delete(
		"type = ? AND affiliate_notification_subject_id = ?",
		t,
		aff_id,
	).Error
}

func DeleteNotificationOfManyType(t []NotificationType, aff_id uint) error {
	return db.Delete(
		"type in ? AND affiliate_notification_subject_id = ?",
		t,
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

func PushThreadReplyNotification(threadId, replyId uint) {
	var authorId uint = 0
	err := db.Model(Thread{}).Where("id = ?", threadId).Select("author_id").Scan(&authorId).Error
	if err != nil {
		logrus.WithError(err).Error("PushThreadReplyNotification: failed to get author id")
	}

	err = PushNotification(&Notification{
		Type:                           NotificationTypeThreadReply,
		UserID:                         authorId,
		AffiliateNotificationSubjectID: &threadId,
	})
	if err != nil {
		logrus.WithError(err).Error("PushThreadReplyNotification: failed to push notification")
	}
}

func PushThreadLikeNotification(threadId, likedUserId uint) {
	var authorId uint = 0
	err := db.Model(Thread{}).Where("id = ?", threadId).Select("author_id").Scan(&authorId).Error
	if err != nil {
		logrus.WithError(err).Error("PushThreadLikeNotification: failed to get author id")
	}

	err = PushNotification(&Notification{
		Type:                           NotificationTypeThreadLike,
		UserID:                         authorId,
		AffiliateNotificationSubjectID: &likedUserId,
	})
	if err != nil {
		logrus.WithError(err).Error("PushThreadLikeNotification: failed to push notification")
	}
}

func PushFollowNotification(followedUserId, followerId uint) {
	err := PushNotification(&Notification{
		Type:                           NotificationTypeFollows,
		UserID:                         followedUserId,
		AffiliateNotificationSubjectID: &followerId,
	})
	if err != nil {
		logrus.WithError(err).Error("PushFollowNotification: failed to push notification")
	}
}
