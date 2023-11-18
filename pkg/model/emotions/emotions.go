package emotions

import "gorm.io/gorm"

type Emotion struct {
	Id       uint   `gorm:"primary_key;auto_increment;not_null"`
	Username string `gorm:"size:256"`
	Emotion  string
	Message  string
	Tip1     string
	Tip2     string
}

func (Emotion) TableName() string {
	return "emotions"
}

type EmotionsData struct {
	Id       int64  `json:"id"`
	Emotion  string `json:"emotion"`
	Message  string `json:"message"`
	Tip1     string `json:"tip1"`
	Tip2     string `json:"tip2"`
	IsCustom bool   `json:"isCustom"`
}

// AddEmotion add new emotion to table
func AddEmotion(db *gorm.DB, t *Emotion) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return tx.Table("emotions").Create(&t).Error
	})
}

// GetEmotions get emotions from table
func GetEmotions(db *gorm.DB, username string) ([]EmotionsData, error) {
	var emotions []Emotion

	if err := db.
		Table("emotions").
		Select("id, emotion, message, tip1, tip2").
		Where("username = ?", username).
		Find(&emotions).
		Error; err != nil {
		return nil, err
	}

	data := GetEmotionsData(emotions)

	return data, nil
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

// GetEmotionsData helper func for constructing emotions response data
func GetEmotionsData(emotions []Emotion) []EmotionsData {
	var data []EmotionsData
	for _, emotion := range emotions {
		data = append(data, EmotionsData{
			Id:       int64(emotion.Id),
			Emotion:  emotion.Emotion,
			Message:  emotion.Message,
			Tip1:     emotion.Tip1,
			Tip2:     emotion.Tip2,
			IsCustom: true,
		})
	}
	return data
}
