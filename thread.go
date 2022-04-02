package models

import (
	"encoding/json"
	"errors"
	"regexp"
	"strconv"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	_ = iota
	ThreadLevelPost
	ThreadLevelComment
	ThreadLevelReply
)

// 一个 Thread 可以是
// 1. 一个帖子。此时 ParentID 和 ReplyID 都为 NULL, Level 为 1.
// 2. 一个楼层。此时 ParentID 为 一楼的 ID，ReplyToID 为 NULL，Level 为 2.
// 3. 一个楼层下的回复。此时 ParentID 为 楼层的 ID，ReplyToID 为回复对象的 ID 或楼层 ID，Level 为 3.
type Thread struct {
	gorm.Model
	Content   postgres.Jsonb `gorm:"type:jsonb;not null" sql:"DEFAULT '{}'::JSONB"`
	LikeCount uint           `gorm:"default:0"`
	Title     string         `gorm:"type:varchar(20);not null"`
	ParentID  *uint          `gorm:"default:null"`
	Parent    *Thread
	ReplyToID *uint `gorm:"deafault:null"`
	ReplyTo   *Thread
	AuthorID  uint `gorm:"not null"`
	Author    *User
	Level     int `gorm:"type:int;default:1"`

	Deleted bool `gorm:"default:false"`
}

func String2Jsonb(s string) postgres.Jsonb {
	return postgres.Jsonb{RawMessage: json.RawMessage(s)}
}

func Jsonb2RawMessage(j postgres.Jsonb) json.RawMessage {
	return j.RawMessage
}

func GetThreadByID(id uint) *Thread {
	thread := Thread{}
	tx := db.Preload("Author").First(&thread, id)
	if tx.Error != nil || thread.Deleted {
		return nil
	} else {
		return &thread
	}
}

func SearchThread(keyword string, uid, page uint) []*Post {
	threads := make([]*Thread, 0)

	db.Preload("Author").Where("title like ? AND level = 1 AND deleted = false", "%"+keyword+"%").Limit(10).Find(&threads)
	res := make([]*Post, len(threads))
	for i, thread := range threads {
		res[i] = ConstructPostObject(*thread, uid)
	}

	ok, _ := regexp.Match("\\d+", []byte(keyword))
	if ok {
		id, _ := strconv.ParseUint(keyword, 10, 32)
		res = append(res, ConstructPostObject(*GetThreadByID(uint(id)), uid))
	}
	return res
}

type Reply struct {
	ID      uint            `json:"id"`
	Content json.RawMessage `json:"content"`
	ReplyTo uint            `json:"reply_to"`
	Author  User            `json:"author"`
}

type Comment struct {
	ID        uint            `json:"id"`
	Content   json.RawMessage `json:"content"`
	Replies   []Reply         `json:"reply_to"`
	Author    User            `json:"author"`
	Likes     uint            `json:"likes"`
	LikedByMe bool            `json:"liked_by_me"`

	commentThreadId uint
}

type Post struct {
	ID      uint                   `json:"id"`
	Title   string                 `json:"title"`
	Content json.RawMessage        `json:"content"`
	Likes   uint                   `json:"likes"`
	Stars   uint                   `json:"stars"`
	Author  map[string]interface{} `json:"author"`

	StaredByMe bool `json:"stared_by_me"`
	LikedByMe  bool `json:"liked_by_me"`

	Comments []Comment `json:"comments"`

	Deleted bool `json:"deleted"`
}

// func postLikedByUser(t *Thread, uid uint) bool {
// 	for _, u := range t.LikedUser {
// 		if u.ID == uid {
// 			return true
// 		}
// 	}
// 	return false
// }

// func postStaredByUser(t *Thread, uid uint) bool {
// 	for _, u := range t.StaredUser {
// 		if u.ID == uid {
// 			return true
// 		}
// 	}
// 	return false
// }

func ConstructPostObject(t Thread, uid uint) *Post {
	threadId := t.ID
	if t.Level != 1 {
		return nil
	}

	res := Post{
		ID:         t.ID,
		Title:      t.Title,
		Content:    Jsonb2RawMessage(t.Content),
		Likes:      FindThreadLikeCount(threadId),
		Stars:      FindThreadStarCount(threadId),
		Author:     t.Author.GetPublicInfomation(),
		StaredByMe: threadStaredByUser(threadId, uid),
		LikedByMe:  threadLikedByUser(threadId, uid),
	}

	// find comments
	commentThreads := make([]Thread, 0)
	tx := db.Preload("Author").Where("parent_id = ? AND deleted = false", threadId).Order("like_count desc, id desc").Find(&commentThreads)

	if tx.Error != nil {
		logrus.Error(tx.Error)
		return nil
	}

	// construct comment
	comments := make([]Comment, len(commentThreads))
	for i, commentThread := range commentThreads {
		comments[i] = Comment{
			ID:              commentThread.ID,
			Content:         Jsonb2RawMessage(commentThread.Content),
			Author:          *commentThread.Author,
			Likes:           FindThreadLikeCount(commentThread.ID),
			LikedByMe:       threadLikedByUser(commentThread.ID, uid),
			commentThreadId: commentThread.ID,
		}
	}

	// find replies for each comment
	for _, comment := range comments {
		replyThreads := make([]Thread, 0)
		tx := db.Preload("Author").Where("parent_id = ? AND deleted = false", comment.commentThreadId).Find(&replyThreads)
		if tx.Error != nil {
			logrus.Error(tx.Error)
			return nil
		}

		comment.Replies = make([]Reply, len(replyThreads))
		for i, reply := range replyThreads {
			comment.Replies[i] = Reply{
				ID:      reply.ID,
				Content: Jsonb2RawMessage(reply.Content),
				ReplyTo: *reply.ReplyToID,
				Author:  *reply.Author,
			}
		}
	}

	res.Comments = comments
	return &res
}

func NewPost(title, content string, author uint) (*Thread, error) {
	thread := Thread{
		Title:    title,
		Content:  String2Jsonb(content),
		AuthorID: author,
		Level:    ThreadLevelPost,
	}

	tx := db.Create(&thread)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &thread, nil
}

var CommentOnThread = ReplyToThread

func ReplyToThread(thread uint, author uint, content string) error {
	commentThread := Thread{
		Content:  String2Jsonb(content),
		ParentID: &thread,
		AuthorID: author,
		Level:    ThreadLevelComment,
	}

	tx := db.Create(&commentThread)
	return tx.Error
}

func ReplyToComment(comment, author uint, content string) error {
	replyThread := Thread{
		Content:  String2Jsonb(content),
		ParentID: &comment,
		AuthorID: author,
		Level:    ThreadLevelReply,
	}

	tx := db.Create(&replyThread)
	return tx.Error
}

func ReplyToReply(comment, author, replyTo uint, content string) error {
	replyThread := Thread{
		Content:   String2Jsonb(content),
		ParentID:  &comment,
		ReplyToID: &replyTo,
		AuthorID:  author,
		Level:     ThreadLevelReply,
	}

	tx := db.Create(&replyThread)
	return tx.Error
}

func LikeThread(threadId uint, userId uint) error {
	thread := Thread{}
	tx := db.First(&thread, threadId)
	if tx.Error != nil || thread.Deleted {
		return NewRequestError("帖子不存在")
	}

	// user := User{}
	// _ = db.First(&user, userId)
	// thread.LikedUser = append(thread.LikedUser, &user)
	// thread.Likes += 1
	// tx = db.Save(thread)

	return CreateThreadLike(threadId, userId)
}

func UnlikeThread(threadId uint, userId uint) error {
	thread := Thread{}
	tx := db.First(&thread, threadId)
	if tx.Error != nil || thread.Deleted {
		return NewRequestError("帖子不存在")
	}

	return DeleteThreadLikeOfThreadForUser(threadId, userId)
}

func StarThread(threadId uint, userId uint) error {
	thread := Thread{}
	tx := db.First(&thread, threadId)
	if tx.Error != nil || thread.Deleted {
		return NewRequestError("帖子不存在")
	}

	if thread.Level != ThreadLevelPost {
		return NewRequestError("不能收藏评论")
	}

	// user := User{}
	// _ = db.First(&user, userId)
	// thread.StaredUser = append(thread.LikedUser, &user)
	// thread.Stars += 1
	// tx = db.Save(thread)

	// return tx.Error

	return CreateThreadStar(threadId, userId)
}

func UnstarThread(threadId uint, userId uint) error {
	thread := Thread{}
	tx := db.First(&thread, threadId)
	if tx.Error != nil {
		return NewRequestError("帖子不存在")
	}

	return DeleteThreadStarOfThreadForUser(threadId, userId)
}

func DeleteThread(id uint) error {
	thread := Thread{}
	tx := db.First(&thread, id)
	if tx.Error != nil {
		return NewRequestError("帖子不存在")
	}

	thread.Deleted = true
	tx = db.Save(thread)

	return tx.Error
}

func GetRandomThreads(count int, uid uint) ([]*Post, error) {
	if count <= 0 {
		return nil, errors.New("count must be greater than 0")
	}
	threads := make([]Thread, 0)
	tx := db.Preload("Author").Where("deleted = false AND level = 1").Order("random()").Limit(count).Order("like_count desc, id desc").Find(&threads)

	posts := make([]*Post, len(threads))
	for i, thread := range threads {
		posts[i] = ConstructPostObject(thread, uid)
	}
	return posts, tx.Error
}

func GetUserReplies(uid uint, page int) ([]Thread, error) {
	if page <= 0 {
		return nil, errors.New("count must be greater than 0")
	}
	threads := make([]Thread, 0)
	tx := db.Preload("Author").Where("deleted = false AND level > 1 AND author_id = ?", uid).Order("id desc").Offset((page - 1) * 10).Limit(10).Find(&threads)

	return threads, tx.Error
}

func GetUserPosts(uid uint, page int) ([]*Post, error) {
	if page <= 0 {
		return nil, errors.New("page must be greater than 0")
	}
	threads := make([]Thread, 0)
	tx := db.Preload("Author").Where("deleted = false AND author_id = ? AND level = 1", uid).Order("id desc").Offset((page - 1) * 10).Limit(10).Find(&threads)

	posts := make([]*Post, len(threads))
	for i, thread := range threads {
		posts[i] = ConstructPostObject(thread, uid)
	}
	return posts, tx.Error
}
