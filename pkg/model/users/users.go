package users

import (
	"github.com/radekkrejcirik01/Koala-backend/pkg/middleware"
	"gorm.io/gorm"
)

type User struct {
	Id           uint   `gorm:"primary_key;auto_increment;not_null" json:"id"`
	Username     string `json:"username"`
	Firstname    string `json:"firstname"`
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
	Username     string `json:"username"`
	Firstname    string `json:"firstname"`
	ProfilePhoto string `json:"profilePhoto"`
}

// CreateUser in users table
func CreateUser(db *gorm.DB, t *User) (string, error) {
	t.Password = middleware.GetHashPassword(t.Password)

	if rows := db.
		Table("users").
		Where("username = ?", t.Username).
		FirstOrCreate(&t).
		RowsAffected; rows == 0 {
		return "User already exists", nil
	}

	return "", nil
}

// LoginUser in users table
func LoginUser(db *gorm.DB, t *Login) error {
	var user User
	t.Password = middleware.GetHashPassword(t.Password)

	if err := db.
		Table("users").
		Where("username = ? AND password = ?", t.Username, t.Password).
		First(&user).
		Error; err != nil {
		return err
	}

	return nil
}

// GetUser from users table
func GetUser(db *gorm.DB, username string) (UserData, error) {
	var user UserData

	err := db.
		Table("users").
		Select("username, firstname, profile_photo").
		Where("username = ?", username).
		Find(&user).
		Error
	if err != nil {
		return UserData{}, err
	}

	return user, nil
}
