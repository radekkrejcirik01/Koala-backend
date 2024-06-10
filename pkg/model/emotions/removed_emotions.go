package emotions

import "gorm.io/gorm"

type RemovedEmotion struct {
	Id        uint `gorm:"primary_key;auto_increment;not_null"`
	Username  string
	EmotionId int
}

func (RemovedEmotion) TableName() string {
	return "removed_emotions"
}

func AddRemovedEmotion(db *gorm.DB, t *RemovedEmotion) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return tx.Table("removed_emotions").Create(&t).Error
	})
}
