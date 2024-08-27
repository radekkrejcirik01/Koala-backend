package users

import (
	"errors"
	"time"

	"github.com/radekkrejcirik01/Koala-backend/pkg/middleware"
	"gorm.io/gorm"
)

type User struct {
	Id           uint   `gorm:"primary_key;auto_increment;not_null" json:"id"`
	Username     string `gorm:"size:256"`
	Name         string `gorm:"size:256"`
	ProfilePhoto string
	Password     string
	LastOnline   int64 `gorm:"autoCreateTime"`
}

func (User) TableName() string {
	return "users"
}

type Login struct {
	Username string
	Password string
}

type Name struct {
	Name string
}

type Password struct {
	OldPassword string
	NewPassword string
}

type UserData struct {
	Id           int64  `json:"id"`
	Username     string `json:"username"`
	Name         string `json:"name"`
	ProfilePhoto string `json:"profilePhoto,omitempty"`
}

type Username struct {
	Username string
}

// CreateUser in users table
func CreateUser(db *gorm.DB, t *User) error {
	t.Password = middleware.GetHashPassword(t.Password)

	return db.Transaction(func(tx *gorm.DB) error {
		if rows := tx.
			Table("users").
			Where("username = ?", t.Username).
			FirstOrCreate(&t).
			RowsAffected; rows == 0 {
			return errors.New("user already exists")
		}
		return nil
	})
}

// LoginUser in users table
func LoginUser(db *gorm.DB, t *Login) error {
	t.Password = middleware.GetHashPassword(t.Password)

	var user User
	return db.
		Table("users").
		Where("username = ? AND password = ?", t.Username, t.Password).
		First(&user).
		Error
}

// GetUser from users table
func GetUser(db *gorm.DB, username string) (UserData, error) {
	var user UserData

	if err := db.
		Table("users").
		Select("id, username, name, profile_photo").
		Where("username = ?", username).
		Find(&user).
		Error; err != nil {
		return UserData{}, err
	}

	return user, nil
}

func CheckUsername(db *gorm.DB, username string) error {
	var user string

	if err := db.
		Table("users").
		Select("username").
		Where("username = ?", username).
		Find(&user).
		Error; err != nil {
		return err
	}

	if len(user) > 0 {
		return errors.New("username is already taken")
	}

	return nil
}

func ChangeName(db *gorm.DB, username string, t *Name) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return tx.
			Table("users").
			Where("username = ?", username).
			Update("name", t.Name).
			Error
	})
}

// ChangePassword change user password in users table
func ChangePassword(db *gorm.DB, username string, t *Password) error {
	oldPassword := middleware.GetHashPassword(t.OldPassword)
	newPassword := middleware.GetHashPassword(t.NewPassword)

	var user User
	err := db.
		Table("users").
		Where("username = ? AND password = ?", username, oldPassword).
		First(&user).
		Error

	if err != nil {
		return errors.New("incorrect old password")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		return tx.
			Table("users").
			Where("id = ?", user.Id).
			Update("password", newPassword).
			Error
	})
}

// UpdateLastOnline in users table
func UpdateLastOnline(db *gorm.DB, username string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return tx.
			Table("users").
			Where("username = ?", username).
			Update("last_online", time.Now().Unix()).
			Error
	})
}

// GetLastOnline from users table
func GetLastOnline(db *gorm.DB, id string) (int64, error) {
	var time int64
	err := db.
		Table("users").
		Select("last_online").
		Where("id = ?", id).
		Find(&time).
		Error
	if err != nil {
		return 0, err
	}
	return time, nil
}
