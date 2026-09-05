package messages

import (
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/notifications"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/recordings"
	"github.com/radekkrejcirik01/Koala-backend/pkg/service"
	"gorm.io/gorm"
)

const EmotionMessageType = "emotion"
const DirectEmotionMessageType = "direct_emotion"
const KudosEmotionMessageType = "kudos"
const MessageType = "message"
const AudioType = "audio"

// Emotion types sent by the client to pick a canned emoji for the message.
// Mirrors the frontend's EmotionType enum.
const FeelingGoodEmotionType = "feeling_good"
const FeelingBadEmotionType = "feeling_bad"
const AngryEmotionType = "angry"
const TiredEmotionType = "tired"
const HeadHurtsEmotionType = "head_hurts"

var emotionEmojis = map[string]string{
	FeelingGoodEmotionType: "😊",
	FeelingBadEmotionType:  "😔",
	AngryEmotionType:       "😠",
	TiredEmotionType:       "😴",
	HeadHurtsEmotionType:   "🤕",
}

// emotionNotificationBodies holds the FCM push body text per emotion type;
// the stored notifications.message stays just the emoji (see emotionEmojis).
var emotionNotificationBodies = map[string]string{
	FeelingGoodEmotionType: "I am feeling good! 😊",
	FeelingBadEmotionType:  "I am not feeling good 😔",
	AngryEmotionType:       "I am feeling angry 😠",
	TiredEmotionType:       "I am feeling tired 😴",
	HeadHurtsEmotionType:   "My head is hurting 🤕",
}

type EmotionMessage struct {
	Ids         []int64
	Message     string
	EmotionType string
}

type Message struct {
	ConversationId int64
	ReceiverId     int64
	Message        string
	ReplyMessage   string
	AudioBuffer    string
}

type User struct {
	Id       int64
	Name     string
	Username string
}

func SendEmotionMessage(db *gorm.DB, t *EmotionMessage, username, messageType string) error {
	var messages []notifications.Notification
	var user User

	if err := db.
		Table("users").
		Select("id, name, username").
		Where("username = ?", username).
		Find(&user).
		Error; err != nil {
		return err
	}

	mType := EmotionMessageType
	if messageType == "direct" {
		mType = DirectEmotionMessageType
	}
	if messageType == "kudos" {
		mType = KudosEmotionMessageType
	}

	message := getEmotionMessage(t.EmotionType, t.Message)
	body := getEmotionNotificationBody(t.EmotionType, t.Message)

	for _, id := range t.Ids {
		messages = append(messages, notifications.Notification{
			SenderId:   user.Id,
			ReceiverId: id,
			Type:       mType,
			Message:    message,
		})
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		return tx.Table("notifications").Create(&messages).Error
	})
	if err != nil {
		return err
	}

	for _, message := range messages {
		err := db.Transaction(func(tx *gorm.DB) error {
			return tx.Table("notifications").
				Where("id = ?", message.Id).
				Update("conversation_id", message.Id).
				Error
		})
		if err != nil {
			return err
		}
	}

	var tokens []string
	tokens, err = service.GetTokensByUserIds(db, t.Ids)
	if err != nil {
		return err
	}

	fcmNotification := service.FcmNotification{
		Title:   user.Name + " is sharing",
		Body:    body,
		Sound:   "default",
		Devices: tokens,
	}

	return service.SendNotification(&fcmNotification)
}

func SendMessage(db *gorm.DB, t *Message, username string) error {
	messageType := MessageType
	var audioMessageUrl string
	var user User

	if err := db.
		Table("users").
		Select("id, name, username").
		Where("username = ?", username).
		Find(&user).
		Error; err != nil {
		return err
	}

	if isAudioMessage(t.AudioBuffer) {
		var err error
		messageType = AudioType

		// Ensure message is emptied when sending voice message
		t.Message = ""

		audioMessageUrl, err = recordings.UploadRecording(t.AudioBuffer, user.Id)
		if err != nil {
			return err
		}
	}

	message := notifications.Notification{
		SenderId:       user.Id,
		ReceiverId:     t.ReceiverId,
		Type:           messageType,
		Message:        t.Message,
		ConversationId: &t.ConversationId,
		ReplyMessage:   &t.ReplyMessage,
		AudioMessage:   &audioMessageUrl,
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		return tx.Table("notifications").Create(&message).Error
	})
	if err != nil {
		return err
	}

	var tokens []string
	tokens, err = service.GetTokensByUserId(db, t.ReceiverId)
	if err != nil {
		return err
	}

	body := t.Message
	if isAudioMessage(audioMessageUrl) {
		body = "🎤 Voice message"
	}

	fcmNotification := service.FcmNotification{
		Title:   user.Name,
		Body:    body,
		Sound:   "default",
		Devices: tokens,
	}

	return service.SendNotification(&fcmNotification)
}

func DeleteMessage(db *gorm.DB, username, id string) error {
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
			Table("notifications").
			Where("id = ? AND sender_id = ?", id, userId).
			Delete(&notifications.Notification{}).
			Error
	})
}

// Check if audio message has length
func isAudioMessage(message string) bool {
	return len(message) > 0
}

// getEmotionMessage resolves the stored/sent message for an emotion share:
// a recognized emotionType maps to its emoji, otherwise falls back to the
// raw message (legacy clients that don't send emotionType).
func getEmotionMessage(emotionType, message string) string {
	if emoji, ok := emotionEmojis[emotionType]; ok {
		return emoji
	}
	return message
}

// getEmotionNotificationBody resolves the FCM push body for an emotion
// share: a recognized emotionType maps to its descriptive text, otherwise
// falls back to the raw message (legacy clients that don't send emotionType).
func getEmotionNotificationBody(emotionType, message string) string {
	if body, ok := emotionNotificationBodies[emotionType]; ok {
		return body
	}
	return message
}
