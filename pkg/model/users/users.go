package users

import (
	"errors"

	"github.com/radekkrejcirik01/Koala-backend/pkg/middleware"
	e "github.com/radekkrejcirik01/Koala-backend/pkg/model/emotions"
	"gorm.io/gorm"
)

type User struct {
	Id           uint   `gorm:"primary_key;auto_increment;not_null" json:"id"`
	Username     string `json:"username"`
	Name         string `json:"name"`
	ProfilePhoto string `json:"profilePhoto"`
	Password     string
}

func (User) TableName() string {
	return "users"
}

type Login struct {
	Username string
	Password string
}

type UserData struct {
	Id       int64  `json:"id,omitempty"`
	Username string `json:"username"`
	Name     string `json:"name"`
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
func GetUser(db *gorm.DB, username string) (UserData, []e.EmotionsData, error) {
	var user UserData
	var emotions []e.Emotion

	if err := db.
		Table("users").
		Select("username, name").
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
