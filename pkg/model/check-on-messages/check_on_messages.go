package checkonmessages

import "gorm.io/gorm"

type CheckOnMessage struct {
	Id      uint `gorm:"primary_key;auto_increment;not_null"`
	UserId  int64
	Message string `gorm:"size:256"`
}

func (CheckOnMessage) TableName() string {
	return "check_on_messages"
}

type CheckOnMessageData struct {
	Id      int64  `json:"id"`
	Message string `json:"message"`
}

func AddCheckOnMessage(db *gorm.DB, t *CheckOnMessage, username string) error {
	var userId int64

	if err := db.
		Table("users").
		Select("id").
		Where("username = ?", username).
		Find(&userId).
		Error; err != nil {
		return err
	}

	t.UserId = userId

	return db.Transaction(func(tx *gorm.DB) error {
		return tx.Table("check_on_messages").Create(&t).Error
	})
}

func GetCheckOnMessages(db *gorm.DB, id string) ([]CheckOnMessageData, error) {
	var replies []CheckOnMessageData

	if err := db.
		Table("check_on_messages").
		Select("id, message").
		Where("user_id = ?", id).
		Find(&replies).
		Error; err != nil {
		return nil, err
	}

	return replies, nil
}

func DeleteCheckOnMessage(db *gorm.DB, username, id string) error {
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
		return tx.
			Table("check_on_messages").
			Where("id = ? AND user_id = ?", id, userId).
			Delete(&CheckOnMessage{}).
			Error
	})
}
