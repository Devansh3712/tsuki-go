package models

import "time"

type Post struct {
	UserId    string
	Id        string
	Body      string `form:"body" binding:"required"`
	Username  string
	Avatar    *string
	CreatedAt time.Time
}

type Comment struct {
	UserId    string
	PostId    string
	Id        string
	Body      string `form:"body" binding:"required"`
	Username  string
	Self      bool
	CreatedAt time.Time
}
