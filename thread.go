package models

import (
	"regexp"
	"strconv"

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
	Content   string  `gorm:"default:{}"`
	Likes     uint    `gorm:"default:0"`
	Stars     uint    `gorm:"default:0"`
	Title     string  `gorm:"type:varchar(20)"`
	ParentID  *uint   `gorm:"parent_id"`
	Parent    *Thread `gorm:"foreignKey:ParentID;default:null;"`
	ReplyToID *uint   `gorm:"reply_to"`
	ReplyTo   *Thread `gorm:"foreignKey:ParentID;default:null;"`
	AuthorID  uint
	Author    *User `gorm:"foreignKey:AuthorID"`
	Level     int   `gorm:"type:tinyint(1);default:1"`

	Deleted bool `gorm:"default:false"`

	LikedUser  []*User `gorm:"many2many:user_liked_thread;"`
	StaredUser []*User `gorm:"many2many:user_stared_thread;"`
}

func GetThreadByID(id uint) *Thread {
	thread := Thread{}
	tx := db.First(&thread, id)
	if tx.Error != nil {
		return nil
	} else {
		return &thread
	}
}

func SearchThread(keyword string, uid, page uint) []*Post {
	res := make([]*Post, 0)

	db.Where("title like ? AND level = 1", "%"+keyword+"%").Limit(10).Find(&res)

	ok, _ := regexp.Match("\\d+", []byte(keyword))
	if ok {
		id, _ := strconv.ParseUint(keyword, 10, 32)
		res = append(res, ConstructPostObject(*GetThreadByID(uint(id)), uid))
	}
	return res
}

type Reply struct {
	ID      uint   `json:"id"`
	Content string `json:"content"`
	ReplyTo uint   `json:"reply_to"`
	Author  User   `json:"author"`
}

type Comment struct {
	ID        uint    `json:"id"`
	Content   string  `json:"content"`
	Replies   []Reply `json:"reply_to"`
	Author    User    `json:"author"`
	Likes     uint    `json:"likes"`
	LikedByMe bool    `json:"liked_by_me"`

	commentThreadId uint
}

type Post struct {
	ID      uint   `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Likes   uint   `json:"likes"`
	Stars   uint   `json:"stars"`
	Author  User   `json:"author"`

	StaredByMe bool `json:"stared_by_me"`
	LikedByMe  bool `json:"liked_by_me"`

	Comments []Comment `json:"comments"`

	Deleted bool `json:"deleted"`
}

func postLikedByUser(t *Thread, uid uint) bool {
	for _, u := range t.LikedUser {
		if u.ID == uid {
			return true
		}
	}
	return false
}

func postStaredByUser(t *Thread, uid uint) bool {
	for _, u := range t.StaredUser {
		if u.ID == uid {
			return true
		}
	}
	return false
}

func ConstructPostObject(t Thread, uid uint) *Post {
	threadId := t.ID
	if t.Level != 1 {
		return nil
	}

	res := Post{
		ID:         t.ID,
		Title:      t.Title,
		Content:    t.Content,
		Likes:      t.Likes,
		Stars:      t.Stars,
		Author:     *t.Author,
		StaredByMe: postStaredByUser(&t, uid),
		LikedByMe:  postLikedByUser(&t, uid),
	}

	// find comments
	commentThreads := make([]Thread, 0)
	tx := db.Where("parent_id = ?", threadId).Order("LikeCount desc").Find(&commentThreads)

	if tx.Error != nil {
		logrus.Error(tx.Error)
		return nil
	}

	// construct comment
	comments := make([]Comment, len(commentThreads))
	for i, commentThread := range commentThreads {
		comments[i] = Comment{
			ID:              commentThread.ID,
			Content:         commentThread.Content,
			Author:          *commentThread.Author,
			Likes:           commentThread.Likes,
			LikedByMe:       postLikedByUser(&commentThread, uid),
			commentThreadId: commentThread.ID,
		}
	}

	// find replies for each comment
	for _, comment := range comments {
		replyThreads := make([]Thread, 0)
		tx := db.Where("parent_id = ?", comment.commentThreadId).Find(&replyThreads)
		if tx.Error != nil {
			logrus.Error(tx.Error)
			return nil
		}

		comment.Replies = make([]Reply, len(replyThreads))
		for i, reply := range replyThreads {
			comment.Replies[i] = Reply{
				ID:      reply.ID,
				Content: reply.Content,
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
		Content:  content,
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
		Content:  content,
		ParentID: &thread,
		AuthorID: author,
		Level:    ThreadLevelComment,
	}

	tx := db.Create(&commentThread)
	return tx.Error
}

func ReplyToComment(thread uint, author uint, replyTo uint, content string) error {
	replyThread := Thread{
		Content:  content,
		ParentID: &thread,
		AuthorID: author,
		Level:    ThreadLevelReply,
	}

	tx := db.Create(&replyThread)
	return tx.Error
}

func LikeThread(threadId uint, userId uint) error {
	thread := Thread{}
	tx := db.First(&thread, threadId)
	if tx.Error != nil {
		return NewRequestError("帖子不存在")
	}

	if postLikedByUser(&thread, userId) {
		return NewRequestError("不能重复点赞")
	}

	user := User{}
	_ = db.First(&user, userId)
	thread.LikedUser = append(thread.LikedUser, &user)
	thread.Likes += 1
	tx = db.Save(thread)

	return tx.Error
}

func StarThread(threadId uint, userId uint) error {
	thread := Thread{}
	tx := db.First(&thread, threadId)
	if tx.Error != nil {
		return NewRequestError("帖子不存在")
	}

	if thread.Level != ThreadLevelPost {
		return NewRequestError("不能收藏评论")
	}

	if postLikedByUser(&thread, userId) {
		return NewRequestError("你已经收藏过该帖子")
	}

	user := User{}
	_ = db.First(&user, userId)
	thread.StaredUser = append(thread.LikedUser, &user)
	thread.Stars += 1
	tx = db.Save(thread)

	return tx.Error
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
