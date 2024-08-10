package emotions

import "gorm.io/gorm"

const AnxietyEmotionType = "anxiety"
const DepressionEmotionType = "depression"
const WellbeingEmotionType = "wellbeing"
const KudosEmotionType = "kudos"

type Emotion struct {
	Id       uint   `gorm:"primary_key;auto_increment;not_null"`
	Username string `gorm:"size:256"`
	Emotion  string
	Message  string `gorm:"size:256"`
	Tip1     string `gorm:"size:256"`
	Tip2     string `gorm:"size:256"`
	Type     string `gorm:"size:20"`
}

func (Emotion) TableName() string {
	return "emotions"
}

type EmotionsData struct {
	Id      int64  `json:"id"`
	Emotion string `json:"emotion"`
	Message string `json:"message"`
	Tip1    string `json:"tip1,omitempty"`
	Tip2    string `json:"tip2,omitempty"`
}

// AddEmotion add new emotion to table
func AddEmotion(db *gorm.DB, t *Emotion) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return tx.Table("emotions").Create(&t).Error
	})
}

func GetEmotionMessages(db *gorm.DB, username, emotionType string) ([]EmotionsData, error) {
	var emotionsMessages []EmotionsData

	if err := db.
		Table("emotions").
		Where("username = ? AND type = ?", username, emotionType).
		Find(&emotionsMessages).
		Error; err != nil {
		return nil, err
	}

	return emotionsMessages, nil
}

// GetEmotions get emotions from table
func GetEmotions(db *gorm.DB, username string) ([]EmotionsData, []int, error) {
	var emotions []EmotionsData
	var removedEmotionsIds []int

	if err := db.
		Table("emotions").
		Select("id, emotion, message, tip1, tip2").
		Where("username = ?", username).
		Find(&emotions).
		Error; err != nil {
		return nil, nil, err
	}

	if err := db.
		Table("removed_emotions").
		Select("emotion_id").
		Where("username = ?", username).
		Find(&removedEmotionsIds).
		Error; err != nil {
		return nil, nil, err
	}

	return emotions, removedEmotionsIds, nil
}

// RemoveEmotion remove emotion from table
func RemoveEmotion(db *gorm.DB, id string, username string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return tx.
			Table("emotions").
			Where("id = ? AND username = ?", id, username).
			Delete(&Emotion{}).
			Error
	})
}
