package emotions

import (
	"strconv"

	"github.com/radekkrejcirik01/Koala-backend/pkg/model/users"
	"gorm.io/gorm"
)

const pageLimit = 20

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

// Recipient is a friend the emotion was shared to, the primary key keeps recipients of one emotion together
type Recipient struct {
	EmotionId int64 `gorm:"primaryKey;autoIncrement:false"`
	UserId    int64 `gorm:"primaryKey;autoIncrement:false;index:idx_emotion_recipients_user"`
}

func (Recipient) TableName() string {
	return "emotion_recipients"
}

type EmotionData struct {
	Id         int64            `json:"id"`
	Emotion    *string          `json:"emotion,omitempty"`
	Message    string           `json:"message"`
	Time       int64            `json:"time"`
	Recipients []users.UserData `gorm:"-" json:"recipients"`
}

type recipientData struct {
	EmotionId    int64
	Id           int64
	Username     string
	Name         string
	ProfilePhoto string
}

// GetEmotions gets user's shared emotions with recipients from emotions table,
// paginated by lastTime and lastId when limit is set, otherwise all emotions in the from and to range
func GetEmotions(db *gorm.DB, username, from, to, lastTime, lastId, limit string) ([]EmotionData, error) {
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
	if lastTime != "" && lastId != "" {
		query = query.Where("(time < ? OR (time = ? AND id < ?))", lastTime, lastTime, lastId)
	}
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil || l <= 0 || l > pageLimit {
			l = pageLimit
		}
		query = query.Limit(l)
	}

	if err := query.
		Order("time DESC, id DESC").
		Find(&emotions).
		Error; err != nil {
		return nil, err
	}

	if len(emotions) == 0 {
		return emotions, nil
	}

	ids := make([]int64, 0, len(emotions))
	for _, emotion := range emotions {
		ids = append(ids, emotion.Id)
	}

	var recipients []recipientData
	if err := db.
		Table("emotion_recipients").
		Select("emotion_recipients.emotion_id, users.id, users.username, users.name, users.profile_photo").
		Joins("JOIN users ON users.id = emotion_recipients.user_id").
		Where("emotion_recipients.emotion_id IN ?", ids).
		Find(&recipients).
		Error; err != nil {
		return nil, err
	}

	recipientsMap := make(map[int64][]users.UserData)
	for _, recipient := range recipients {
		recipientsMap[recipient.EmotionId] = append(recipientsMap[recipient.EmotionId], users.UserData{
			Id:           recipient.Id,
			Username:     recipient.Username,
			Name:         recipient.Name,
			ProfilePhoto: recipient.ProfilePhoto,
		})
	}

	// Emotions shared before recipients were saved have an empty list
	for i := range emotions {
		emotions[i].Recipients = recipientsMap[emotions[i].Id]
		if emotions[i].Recipients == nil {
			emotions[i].Recipients = []users.UserData{}
		}
	}

	return emotions, nil
}

// DeleteUserRecipients deletes recipients of user's emotions and rows where user is a recipient
func DeleteUserRecipients(tx *gorm.DB, userId int64) error {
	emotionIds := tx.
		Table("emotions").
		Select("id").
		Where("user_id = ?", userId)

	return tx.
		Table("emotion_recipients").
		Where("emotion_id IN (?) OR user_id = ?", emotionIds, userId).
		Delete(&Recipient{}).
		Error
}
