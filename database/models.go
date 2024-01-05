package database

import "time"

type User struct {
	ID               int
	Username         string
	Email            string
	Password         string
	RegistrationDate string
}

type Post struct {
	ID           int
	UserID       int
	Title        string
	Content      string
	Category     string
	Comments     []Comment
	LikeCount    int
	DislikeCount int
	CreationDate time.Time
	Username     string
}
type Comment struct {
	ID           int
	UserID       int
	PostID       int
	Username     string
	Content      string
	LikeCount    int
	DislikeCount int
	CreationDate time.Time
}
type Category struct {
	ID   int
	Name string
}
type PostLike struct {
	ID     int
	UserID int
	PostID int
	Like   int // 1 for like, 0 for dislike
}

type CommentLike struct {
	ID        int
	UserID    int
	CommentID int
	Like      int // 1 for like, 0 for dislike
}

type PageData struct {
	IsAuthenticated bool
	Username        string
	Posts           []Post
}
