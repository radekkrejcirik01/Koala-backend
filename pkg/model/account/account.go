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
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Table("users").
			Where("username = ?", username).
			Delete(&users.User{}).
			Error; err != nil {
			return err
		}

		if err := tx.
			Table("devices").
			Where("username = ?", username).
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
			Where("sender = ? OR receiver = ?", username, username).
			Delete(&notifications.Notification{}).
			Error; err != nil {
			return err
		}
		return nil
	})
}
