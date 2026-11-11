package emotions

import (
	"gorm.io/gorm"
)

type Emotion struct {
	Id      uint    `gorm:"primary_key;auto_increment;not_null"`
	UserId  int64   `gorm:"index:idx_emotions_user_time"`
	Emotion *string `gorm:"size:64"`
	Message string  `gorm:"size:512"`
	Time    int64   `gorm:"autoCreateTime;index:idx_emotions_user_time"`
}

func (Emotion) TableName() string {
	return "emotions"
}

type EmotionData struct {
	Id      int64   `json:"id"`
	Emotion *string `json:"emotion,omitempty"`
	Message string  `json:"message"`
	Time    int64   `json:"time"`
}

// GetEmotions gets user's shared emotions from emotions table
func GetEmotions(db *gorm.DB, username, from, to string) ([]EmotionData, error) {
	var emotions []EmotionData
	var userId int64

	if err := db.
		Table("users").
		Select("id").
		Where("username = ?", username).
		Find(&userId).
		Error; err != nil {
		return nil, err
	}

	query := db.
		Table("emotions").
		Select("id, emotion, message, time").
		Where("user_id = ?", userId)

	if from != "" {
		query = query.Where("time >= ?", from)
	}
	if to != "" {
		query = query.Where("time < ?", to)
	}

	if err := query.
		Order("time DESC").
		Find(&emotions).
		Error; err != nil {
		return nil, err
	}

	return emotions, nil
}
