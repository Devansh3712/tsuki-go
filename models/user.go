package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Email     *string `form:"email" binding:"required"`
	Username  string  `form:"username" binding:"required"`
	Password  string  `form:"password" binding:"required"`
	Id        string
	Verified  bool
	Avatar    *string
	CreatedAt time.Time
}

type DiscordUser struct {
	Email     *string `json:"email"`
	Username  string  `json:"username"`
	Verified  bool
	Avatar    *string `json:"avatar"`
	DiscordId string  `json:"id"`
}

type GitHubUser struct {
	Email    *string `json:"email"`
	Username string  `json:"login"`
	Verified bool
	Avatar   *string `json:"avatar_url"`
}

type GoogleUser struct {
	Email    string  `json:"email"`
	Username string  `json:"given_name"`
	Avatar   *string `json:"picture"`
	Verified bool    `json:"email_verified"`
}

type Login struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

func (u *User) HashPassword() error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
