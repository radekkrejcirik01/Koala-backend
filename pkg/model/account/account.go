package account

import (
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/devices"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/invites"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/notifications"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/users"
	"gorm.io/gorm"
)

// DeleteAccount delete user from tables
func DeleteAccount(db *gorm.DB, username string) error {
	var userId int64

	if err := db.
		Table("users").
		Select("id").
		Where("username = ?", username).
		Find(&userId).
		Error; err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Table("users").
			Where("id = ?", userId).
			Delete(&users.User{}).
			Error; err != nil {
			return err
		}

		if err := tx.
			Table("devices").
			Where("user_id = ?", userId).
			Delete(&devices.Device{}).
			Error; err != nil {
			return err
		}

		if err := tx.
			Table("invites").
			Where("sender = ? OR receiver = ?", username, username).
			Delete(&invites.Invite{}).
			Error; err != nil {
			return err
		}

		if err := tx.
			Table("notifications").
			Where("sender = ? OR receiver = ? OR sender_id = ? OR receiver_id = ?",
				username, username, userId, userId).
			Delete(&notifications.Notification{}).
			Error; err != nil {
			return err
		}
		return nil
	})
}
