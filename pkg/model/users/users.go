package users

import (
	"errors"
	"time"

	"github.com/radekkrejcirik01/Koala-backend/pkg/middleware"
	e "github.com/radekkrejcirik01/Koala-backend/pkg/model/emotions"
	"github.com/radekkrejcirik01/Koala-backend/pkg/service"
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

type Password struct {
	OldPassword string
	NewPassword string
}

type UserData struct {
	Id       int64  `json:"id,omitempty"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

// CreateUser in users table
func CreateUser(db *gorm.DB, t *User) error {
	t.Password = middleware.GetHashPassword(t.Password)

	err := db.Transaction(func(tx *gorm.DB) error {
		if rows := tx.
			Table("users").
			Where("username = ?", t.Username).
			FirstOrCreate(&t).
			RowsAffected; rows == 0 {
			return errors.New("user already exists")
		}
		return nil
	})
	if err != nil {
		return err
	}

	var tokens []string
	tokens, err = service.GetTokensByUserId(db, 123)
	if err != nil {
		return err
	}

	fcmNotification := service.FcmNotification{
		Body:    "New user joined Koala!",
		Sound:   "default",
		Devices: tokens,
	}

	return service.SendNotification(&fcmNotification)
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
func GetUser(db *gorm.DB, username string) (UserData, []e.EmotionsData, error) {
	var user UserData
	var emotions []e.Emotion

	if err := db.
		Table("users").
		Select("id, username, name").
		Where("username = ?", username).
		Find(&user).
		Error; err != nil {
		return UserData{}, nil, err
	}

	if err := db.
		Table("emotions").
		Select("id, emotion, message, tip1, tip2").
		Where("username = ?", username).
		Find(&emotions).
		Error; err != nil {
		return UserData{}, nil, err
	}

	emotionsData := e.GetEmotionsData(emotions)

	return user, emotionsData, nil
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
