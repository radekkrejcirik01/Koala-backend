package devices

import "gorm.io/gorm"

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

func SaveDevice(db *gorm.DB, t *Device) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return tx.Table("devices").FirstOrCreate(t, t).Error
	})
}

func DeleteDevice(db *gorm.DB, username string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return tx.
			Table("devices").
			Where("username = ?", username).
			Delete(&Device{}).
			Error
	})
}
