package messages

import (
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/notifications"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/recordings"
	"github.com/radekkrejcirik01/Koala-backend/pkg/service"
	"gorm.io/gorm"
)

const EmotionMessageType = "emotion"
const MessageType = "message"
const AudioType = "audio"
const StatusReplyType = "status_reply"

type EmotionMessage struct {
	Ids     []int64
	Message string
}

type Message struct {
	ConversationId int64
	ReceiverId     int64
	Message        string
	ReplyMessage   string
	AudioBuffer    string
}

type StatusReplyMessage struct {
	ReceiverId      int64
	Message         string
	ReplyExpression string
}

type User struct {
	Id   int64
	Name string
}

func SendEmotionMessage(db *gorm.DB, t *EmotionMessage, username string) error {
	var messages []notifications.Notification
	var user User

	if err := db.
		Table("users").
		Select("id, name").
		Where("username = ?", username).
		Find(&user).
		Error; err != nil {
		return err
	}

	for _, id := range t.Ids {
		messages = append(messages, notifications.Notification{
			SenderId:   user.Id,
			ReceiverId: id,
			Type:       EmotionMessageType,
			Message:    t.Message,
		})
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		return tx.Table("notifications").Create(&messages).Error
	})
	if err != nil {
		return err
	}

	var tokens []string
	tokens, err = service.GetTokensByUserIds(db, t.Ids)
	if err != nil {
		return err
	}

	fcmNotification := service.FcmNotification{
		Title:   user.Name + " is sharing",
		Body:    t.Message,
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
		Select("id, name").
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

func SendStatusReplyMessage(db *gorm.DB, t *StatusReplyMessage, username string) error {
	var user User

	if err := db.
		Table("users").
		Select("id, name").
		Where("username = ?", username).
		Find(&user).
		Error; err != nil {
		return err
	}

	message := notifications.Notification{
		SenderId:     user.Id,
		ReceiverId:   t.ReceiverId,
		Message:      t.Message,
		Type:         StatusReplyType,
		ReplyMessage: &t.ReplyExpression,
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		return tx.Table("notifications").Create(&message).Error
	})
	if err != nil {
		return err
	}

	tokens, err := service.GetTokensByUserId(db, t.ReceiverId)
	if err != nil {
		return err
	}

	fcmNotification := service.FcmNotification{
		Body:    user.Name + " is replying to your status: " + t.Message,
		Sound:   "default",
		Devices: tokens,
	}

	return service.SendNotification(&fcmNotification)
}

// Check if audio message has length
func isAudioMessage(message string) bool {
	return len(message) > 0
}
