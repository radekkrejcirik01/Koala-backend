package devices

import (
	"gorm.io/gorm"
)

type Device struct {
	Id          uint   `gorm:"primary_key;auto_increment;not_null"`
	Username    string `gorm:"size:256"`
	UserId      int64
	DeviceToken string
	Platform    string
}

func (Device) TableName() string {
	return "devices"
}

// SaveDevice add new device to devices table
func SaveDevice(db *gorm.DB, t *Device) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return tx.
			Table("devices").
			Where("username = ? AND device_token = ? AND user_id = ?",
				t.Username, t.DeviceToken, t.UserId).
			FirstOrCreate(&t).
			Error
	})
}

// DeleteDevice delete all devices from devices table
func DeleteDevice(db *gorm.DB, username string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return tx.
			Table("devices").
			Where("username = ?", username).
			Delete(&Device{}).
			Error
	})
}
