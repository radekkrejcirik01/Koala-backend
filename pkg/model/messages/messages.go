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
const StatusReplyType = "status_reply"
const CheckOnType = "check_on"

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

type LastSharedMessage struct {
	Id      int64  `json:"id"`
	Message string `json:"message"`
	Tip1    string `json:"tip1"`
	Tip2    string `json:"tip2"`
	Type    string `json:"type"`
}

type StatusReplyMessage struct {
	ReceiverId      int64
	Message         string
	ReplyExpression string
}

type CheckOnMessage struct {
	Ids     []int64
	Message string
}

type User struct {
	Id   int64
	Name string
}

func SendEmotionMessage(db *gorm.DB, t *EmotionMessage, username, messageType string) error {
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

	mType := EmotionMessageType
	if messageType == "direct" {
		mType = DirectEmotionMessageType
	}
	if messageType == "kudos" {
		mType = KudosEmotionMessageType
	}

	for _, id := range t.Ids {
		messages = append(messages, notifications.Notification{
			SenderId:   user.Id,
			ReceiverId: id,
			Type:       mType,
			Message:    t.Message,
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
		Title:   "💬 " + user.Name,
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
		Title:   "💬 " + user.Name,
		Body:    body,
		Sound:   "default",
		Devices: tokens,
	}

	return service.SendNotification(&fcmNotification)
}

func GetLastSharedMessage(db *gorm.DB, username string) (LastSharedMessage, error) {
	var lastSharedMessage LastSharedMessage
	var userId int64
	var message string

	types := []string{
		EmotionMessageType,
		KudosEmotionMessageType,
		DirectEmotionMessageType,
	}

	if err := db.
		Table("users").
		Select("id").
		Where("username = ?", username).
		Find(&userId).
		Error; err != nil {
		return LastSharedMessage{}, err
	}

	if err := db.
		Table("notifications").
		Select("message").
		Where("sender_id = ? AND type IN ?", userId, types).
		Order("id DESC").
		Limit(1).
		Find(&message).
		Error; err != nil {
		return LastSharedMessage{}, err
	}

	if err := db.
		Table("emotions").
		Where("username = ? AND message = ?", username, message).
		Order("id DESC").
		Limit(1).
		Find(&lastSharedMessage).
		Error; err != nil {
		return LastSharedMessage{}, err
	}

	if len(lastSharedMessage.Message) == 0 {
		lastSharedMessage.Message = message
	}

	return lastSharedMessage, nil
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

	err = db.Transaction(func(tx *gorm.DB) error {
		return tx.Table("notifications").
			Where("id = ?", message.Id).
			Update("conversation_id", message.Id).
			Error
	})
	if err != nil {
		return err
	}

	tokens, err := service.GetTokensByUserId(db, t.ReceiverId)
	if err != nil {
		return err
	}

	fcmNotification := service.FcmNotification{
		Title:   "Status reply 💬",
		Body:    user.Name + ": " + t.Message,
		Sound:   "default",
		Devices: tokens,
	}

	return service.SendNotification(&fcmNotification)
}

func SendCheckOnMessage(db *gorm.DB, t *CheckOnMessage, username string) error {
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
			Type:       CheckOnType,
			Message:    t.Message,
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
		Title:   "💬 " + user.Name,
		Body:    t.Message,
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
