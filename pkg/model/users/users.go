package users

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/radekkrejcirik01/Koala-backend/pkg/middleware"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/codes"
	"github.com/radekkrejcirik01/Koala-backend/pkg/service"
	"gorm.io/gorm"
)

type User struct {
	Id           uint   `gorm:"primary_key;auto_increment;not_null" json:"id"`
	Username     string `gorm:"size:256;index"`
	Name         string `gorm:"size:256"`
	ProfilePhoto string
	Password     string
	LastOnline   int64 `gorm:"autoCreateTime"`
	// Email is set only once it's verified, it's used for password reset
	Email        string `gorm:"size:256"`
	TokenVersion int64  `gorm:"not null;default:0"`
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
	Email        string `json:"email,omitempty"`
}

type Username struct {
	Username string
}

type Email struct {
	Email string
}

type EmailCode struct {
	Code string
}

type PasswordReset struct {
	Username    string
	Code        string
	NewPassword string
}

var ErrInvalidEmail = errors.New("invalid email")

// CreateUser in users table
func CreateUser(db *gorm.DB, t *User) error {
	password, err := middleware.HashPassword(t.Password)
	if err != nil {
		return err
	}

	// Only these come from the request, email and token version can't be set at sign up
	user := User{
		Username:     t.Username,
		Name:         t.Name,
		ProfilePhoto: t.ProfilePhoto,
		Password:     password,
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if rows := tx.
			Table("users").
			Where("username = ?", user.Username).
			FirstOrCreate(&user).
			RowsAffected; rows == 0 {
			return errors.New("user already exists")
		}
		return nil
	})
}

// LoginUser in users table, returns the token version for the new token
func LoginUser(db *gorm.DB, t *Login) (int64, error) {
	var user User
	if err := db.
		Table("users").
		Select("id, password, token_version").
		Where("username = ?", t.Username).
		Take(&user).
		Error; err != nil {
		return 0, err
	}

	ok, legacy := middleware.CheckPassword(user.Password, t.Password)
	if !ok {
		// Same error as an unknown username
		return 0, gorm.ErrRecordNotFound
	}

	// Legacy sha512 hash is replaced now that the password is known,
	// login doesn't fail when it can't be, it's tried again next time
	if legacy {
		if hash, err := middleware.HashPassword(t.Password); err == nil {
			db.
				Table("users").
				Where("id = ?", user.Id).
				Update("password", hash)
		}
	}

	return user.TokenVersion, nil
}

// GetUser from users table
func GetUser(db *gorm.DB, username string) (UserData, error) {
	var user UserData

	if err := db.
		Table("users").
		Select("id, username, name, profile_photo, email").
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
	return db.
		Table("users").
		Where("username = ?", username).
		Update("name", t.Name).
		Error
}

// ChangePassword change user password in users table
func ChangePassword(db *gorm.DB, username string, t *Password) error {
	var user User
	if err := db.
		Table("users").
		Select("id, password").
		Where("username = ?", username).
		Take(&user).
		Error; err != nil {
		return err
	}

	if ok, _ := middleware.CheckPassword(user.Password, t.OldPassword); !ok {
		return errors.New("incorrect old password")
	}

	newPassword, err := middleware.HashPassword(t.NewPassword)
	if err != nil {
		return err
	}

	return db.
		Table("users").
		Where("id = ?", user.Id).
		Update("password", newPassword).
		Error
}

// ChangeEmail sends a code to the new email, it's saved once VerifyEmail gets the code
func ChangeEmail(db *gorm.DB, username string, t *Email) error {
	address, err := mail.ParseAddress(strings.TrimSpace(t.Email))
	if err != nil {
		return ErrInvalidEmail
	}

	var id int64
	if err := db.
		Table("users").
		Select("id").
		Where("username = ?", username).
		Take(&id).
		Error; err != nil {
		return err
	}

	code, err := codes.Create(db, id, codes.PurposeEmail, address.Address)
	if err != nil {
		return err
	}

	return service.SendEmailVerificationCode(address.Address, code)
}

// VerifyEmail saves the email the code was sent to
func VerifyEmail(db *gorm.DB, username string, t *EmailCode) error {
	var id int64
	if err := db.
		Table("users").
		Select("id").
		Where("username = ?", username).
		Take(&id).
		Error; err != nil {
		return err
	}

	code, err := codes.Verify(db, id, codes.PurposeEmail, t.Code)
	if err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := codes.Delete(tx, id, codes.PurposeEmail); err != nil {
			return err
		}

		return tx.
			Table("users").
			Where("id = ?", id).
			Update("email", code.Email).
			Error
	})
}

// RequestPasswordReset emails a reset code to the user's verified email.
// It succeeds also when the user doesn't exist, has no email or got a code recently,
// so the answer doesn't tell which usernames exist.
func RequestPasswordReset(db *gorm.DB, t *Username) error {
	var user User
	if err := db.
		Table("users").
		Select("id, email").
		Where("username = ?", t.Username).
		Find(&user).
		Error; err != nil {
		return err
	}

	if user.Id == 0 || user.Email == "" {
		return nil
	}

	code, err := codes.Create(db, int64(user.Id), codes.PurposePasswordReset, user.Email)
	if errors.Is(err, codes.ErrTooSoon) {
		return nil
	}
	if err != nil {
		return err
	}

	return service.SendPasswordResetCode(user.Email, t.Username, code)
}

// ResetPassword sets the new password and revokes all tokens,
// returns the new token version for the new token
func ResetPassword(db *gorm.DB, t *PasswordReset) (int64, error) {
	if t.NewPassword == "" {
		return 0, errors.New("new password is empty")
	}

	var id int64
	if err := db.
		Table("users").
		Select("id").
		Where("username = ?", t.Username).
		Find(&id).
		Error; err != nil {
		return 0, err
	}

	if id == 0 {
		return 0, codes.ErrInvalid
	}

	if _, err := codes.Verify(db, id, codes.PurposePasswordReset, t.Code); err != nil {
		return 0, err
	}

	password, err := middleware.HashPassword(t.NewPassword)
	if err != nil {
		return 0, err
	}

	var tokenVersion int64
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := codes.Delete(tx, id, codes.PurposePasswordReset); err != nil {
			return err
		}

		if err := tx.
			Table("users").
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"password":      password,
				"token_version": gorm.Expr("token_version + 1"),
			}).
			Error; err != nil {
			return err
		}

		return tx.
			Table("users").
			Select("token_version").
			Where("id = ?", id).
			Take(&tokenVersion).
			Error
	})

	return tokenVersion, err
}

// UpdateLastOnline in users table
func UpdateLastOnline(db *gorm.DB, username string) error {
	return db.
		Table("users").
		Where("username = ?", username).
		Update("last_online", time.Now().Unix()).
		Error
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
