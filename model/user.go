package model

import (
	"errors"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

const (
	UserTypeStudent = 0
	UserTypeTeacher = 1
	UserTypeAdmin   = 2
)

var DefaultUser = new(User)

var ErrUserNameExisted = errors.New("用户名已存在")

type User struct {
	ID           uint `gorm:"primary_key"`
	CreatedAt    time.Time
	Username     string `gorm:"not null;unique;size:20"`
	PasswordHash []byte `gorm:"type:BINARY(60);not null"`
	UserType     int
}

func (u *User) SetPassword(password string) (err error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	u.PasswordHash = hash
	return
}

func (u *User) Create() (err error) {
	err = db.First(u, "username = ?", u.Username).Error
	if err == nil {
		return ErrUserNameExisted
	}
	if err != gorm.ErrRecordNotFound {
		return
	}

	err = db.Create(u).Error
	return
}

func (u *User) FirstByUsername() (err error) {
	err = db.First(u, "username = ?", u.Username).Error
	return
}

func (u *User) DeleteById() (err error) {
	err = db.Delete(u, u.ID).Error
	return
}

func (u *User) Save() (err error) {
	err = db.Save(u).Error
	return
}

func (u *User) UpdatePassword() (err error) {
	err = db.Model(u).Update("password_hash", u.PasswordHash).Error
	return
}

func (u *User) ComparePassword(password string) (err error) {
	err = bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password))
	return
}

func (u *User) GetUsers() (users []User, err error) {
	err = db.Find(&users).Error
	return
}

func (u *User) Count() (count uint, err error) {
	err = db.Model(u).Count(&count).Error
	return
}

func (u *User) CreateDefaultAdminUser() (err error) {
	u = &User{Username: "admin", UserType: UserTypeAdmin}
	err = u.SetPassword("admin123")
	if err != nil {
		return
	}

	err = u.Create()
	return
}

func (u *User) AfterDelete(tx *gorm.DB) (err error) {
	err = DefaultLoginToken.LogoutAll(u.ID)
	if err != nil {
		log.Printf("fail to LogoutAll: %v", err)
		tx.Rollback()
		return
	}

	err = DefaultProject.DeleteByUserId(u.ID)
	if err != nil {
		log.Printf("fail to delete project by user id: %v", err)
		tx.Rollback()
		return
	}
	return
}

func (u *User) AfterUpdate(tx *gorm.DB) (err error) {
	err = DefaultLoginToken.LogoutAll(u.ID)
	if err != nil {
		log.Printf("fail to LogoutAll: %v", err)
		return
	}
	return
}
