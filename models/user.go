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
