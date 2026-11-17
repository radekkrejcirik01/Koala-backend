package codes

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"
)

const (
	PurposeEmail         = "email"
	PurposePasswordReset = "password-reset"

	expiration  = 15 * 60
	resendDelay = 60
	maxAttempts = 5
)

var (
	ErrTooSoon = errors.New("code was sent recently, try again in a minute")
	ErrInvalid = errors.New("invalid or expired code")
)

// Code is a one-time code sent by email, only its hash is stored.
// A user has at most one code per purpose, a new one replaces the old one.
type Code struct {
	UserId    int64  `gorm:"primaryKey;autoIncrement:false"`
	Purpose   string `gorm:"primaryKey;size:32"`
	Hash      string `gorm:"size:64"`
	Email     string `gorm:"size:256"`
	Attempts  int
	CreatedAt int64
	ExpiresAt int64
}

func (Code) TableName() string {
	return "codes"
}

// Create replaces the user's code for the purpose and returns the new code
func Create(db *gorm.DB, userId int64, purpose, email string) (string, error) {
	var createdAt int64
	if err := db.
		Table("codes").
		Select("created_at").
		Where("user_id = ? AND purpose = ?", userId, purpose).
		Find(&createdAt).
		Error; err != nil {
		return "", err
	}

	now := time.Now().Unix()
	if now-createdAt < resendDelay {
		return "", ErrTooSoon
	}

	code, err := generate()
	if err != nil {
		return "", err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Table("codes").
			Where("user_id = ? AND purpose = ?", userId, purpose).
			Delete(&Code{}).
			Error; err != nil {
			return err
		}

		return tx.
			Table("codes").
			Create(&Code{
				UserId:    userId,
				Purpose:   purpose,
				Hash:      hash(code),
				Email:     email,
				CreatedAt: now,
				ExpiresAt: now + expiration,
			}).
			Error
	})

	return code, err
}

// Verify checks the code without using it up, call Delete once the change is made
func Verify(db *gorm.DB, userId int64, purpose, code string) (Code, error) {
	// The attempt is counted before comparing, so parallel requests can't try more codes than allowed
	result := db.
		Table("codes").
		Where("user_id = ? AND purpose = ? AND attempts < ? AND expires_at > ?",
			userId, purpose, maxAttempts, time.Now().Unix()).
		Update("attempts", gorm.Expr("attempts + 1"))
	if result.Error != nil {
		return Code{}, result.Error
	}
	if result.RowsAffected == 0 {
		return Code{}, ErrInvalid
	}

	var c Code
	if err := db.
		Table("codes").
		Where("user_id = ? AND purpose = ?", userId, purpose).
		Take(&c).
		Error; err != nil {
		return Code{}, err
	}

	if subtle.ConstantTimeCompare([]byte(c.Hash), []byte(hash(code))) != 1 {
		return Code{}, ErrInvalid
	}

	return c, nil
}

// Delete uses up the code, it fails when a parallel request already used it
func Delete(tx *gorm.DB, userId int64, purpose string) error {
	result := tx.
		Table("codes").
		Where("user_id = ? AND purpose = ?", userId, purpose).
		Delete(&Code{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrInvalid
	}

	return nil
}

func DeleteUserCodes(tx *gorm.DB, userId int64) error {
	return tx.
		Table("codes").
		Where("user_id = ?", userId).
		Delete(&Code{}).
		Error
}

func generate() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n.Int64()), nil
}

func hash(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}
