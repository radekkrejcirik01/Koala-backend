package replies

import "gorm.io/gorm"

type Reply struct {
	Id      uint `gorm:"primary_key;auto_increment;not_null"`
	UserId  int64
	Message string `gorm:"size:256"`
}

func (Reply) TableName() string {
	return "replies"
}

type ReplyData struct {
	Id      int64  `json:"id"`
	Message string `json:"message"`
}

func AddReply(db *gorm.DB, t *Reply, username string) error {
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
		return tx.Table("replies").Create(&t).Error
	})
}

func GetReplies(db *gorm.DB, id string) ([]ReplyData, error) {
	var replies []ReplyData

	if err := db.
		Table("replies").
		Select("id, message").
		Where("user_id = ?", id).
		Find(&replies).
		Error; err != nil {
		return nil, err
	}

	return replies, nil
}

func DeleteReply(db *gorm.DB, username, id string) error {
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
			Table("replies").
			Where("id = ? AND user_id = ?", id, userId).
			Delete(&Reply{}).
			Error
	})
}
