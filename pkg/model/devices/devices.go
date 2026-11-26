package devices

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Device struct {
	Id          uint   `gorm:"primary_key;auto_increment;not_null"`
	UserId      int64  `gorm:"index"`
	DeviceToken string `gorm:"size:255;uniqueIndex"`
	Platform    string
	Version     string
}

func (Device) TableName() string {
	return "devices"
}

type DeleteRequest struct {
	DeviceToken string
}

// SaveDevice saves the device's token for the user, a token has one row,
// registering it again moves it to the user and updates platform and version
func SaveDevice(db *gorm.DB, username string, t *Device) error {
	if t.DeviceToken == "" {
		return errors.New("device token is required")
	}

	userId, err := getUserId(db, username)
	if err != nil {
		return err
	}

	device := Device{
		UserId:      userId,
		DeviceToken: t.DeviceToken,
		Platform:    t.Platform,
		Version:     t.Version,
	}

	return db.
		Table("devices").
		Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"user_id", "platform", "version"}),
		}).
		Create(&device).
		Error
}

// DeleteDevice deletes the user's device with the token,
// without a token all user's devices are deleted, as old app versions don't send it
func DeleteDevice(db *gorm.DB, username, deviceToken string) error {
	userId, err := getUserId(db, username)
	if err != nil {
		return err
	}

	query := db.
		Table("devices").
		Where("user_id = ?", userId)

	if deviceToken != "" {
		query = query.Where("device_token = ?", deviceToken)
	}

	return query.Delete(&Device{}).Error
}

func getUserId(db *gorm.DB, username string) (int64, error) {
	var userId int64
	err := db.
		Table("users").
		Select("id").
		Where("username = ?", username).
		Find(&userId).
		Error
	if err == nil && userId == 0 {
		err = errors.New("user not found")
	}

	return userId, err
}
