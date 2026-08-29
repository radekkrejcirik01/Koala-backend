package notifications

import (
	"fmt"

	"github.com/radekkrejcirik01/Koala-backend/pkg/model/users"
	"github.com/radekkrejcirik01/Koala-backend/pkg/service"
	"gorm.io/gorm"
)

type Notification struct {
	Id             uint   `gorm:"primary_key;auto_increment;not_null"`
	Sender         string `gorm:"size:256"`
	SenderId       int64
	Receiver       string `gorm:"size:256"`
	ReceiverId     int64
	Type           string `gorm:"size:20"`
	Message        string `gorm:"size:512"`
	Time           int64  `gorm:"autoCreateTime"`
	Seen           int    `gorm:"default:0"`
	ConversationId *int64
	ReplyMessage   *string `gorm:"size:512"`
	AudioMessage   *string `gorm:"size:512"`
	Reaction       *string `gorm:"size:10"`
}

func (Notification) TableName() string {
	return "notifications"
}

type NotificationData struct {
	Id             int64  `json:"id"`
	SenderId       int64  `json:"senderId"`
	Sender         string `json:"sender"`
	Name           string `json:"name"`
	ProfilePhoto   string `json:"profilePhoto,omitempty"`
	Type           string `json:"type"`
	Message        string `json:"message"`
	Time           int64  `json:"time"`
	Seen           int    `json:"seen"`
	ConversationId *int64 `json:"conversationId,omitempty"`
}

type EmotionData struct {
	Id      int64
	Message string
}

type ExpressionData struct {
	Id           int64
	ReplyMessage string
}

type Conversation struct {
	Id           int64  `json:"id"`
	SenderId     int64  `json:"senderId"`
	Sender       string `json:"sender"`
	Receiver     string `json:"receiver"`
	Message      string `json:"message"`
	Type         string `json:"type"`
	Time         int64  `json:"time"`
	ReplyMessage string `json:"replyMessage"`
	AudioMessage string `json:"audioMessage"`
	Reaction     string `json:"reaction,omitempty"`
}

type UpdateReactionRequest struct {
	Reaction *string
}

type TrackData struct {
	Id             int      `json:"id"`
	ReceiversNames []string `json:"receiversNames"`
	Message        string   `json:"message"`
	Time           int64    `json:"time"`
}

// GetNotifications gets notifications from notifications table
func GetNotifications(db *gorm.DB, username string, lastId string) ([]NotificationData, error) {
	var notifications []Notification
	var usersData []users.UserData

	var idCondition string
	if lastId != "" {
		idCondition = fmt.Sprintf("id < %s AND ", lastId)
	}

	var userId int64
	if err := db.
		Table("users").
		Select("id").
		Where("username = ?", username).
		Find(&userId).
		Error; err != nil {
		return nil, err
	}

	if err := db.
		Table("notifications").
		Where(`(sender_id = ? OR receiver_id = ?) AND `+idCondition+`id IN (SELECT
			MAX(id)
			FROM notifications
		WHERE
			(sender_id = ? AND (type = 'message' OR type = 'audio')) OR receiver_id = ?
		GROUP BY
			conversation_id)`,
			userId, userId, userId, userId).
		Order("id DESC").
		Limit(20).
		Find(&notifications).
		Error; err != nil {
		return nil, err
	}

	userIds := getUserIdsFromNotifications(notifications, userId)

	if err := db.
		Table("users").
		Where("id IN ?", userIds).
		Find(&usersData).
		Error; err != nil {
		return nil, err
	}

	var notificationsData []NotificationData
	for _, notification := range notifications {
		user := getNotificationUser(usersData, notification.SenderId, notification.ReceiverId)

		seen := notification.Seen
		if notification.SenderId == userId {
			seen = 1
		}

		notificationsData = append(notificationsData, NotificationData{
			Id:             int64(notification.Id),
			SenderId:       user.Id,
			Sender:         user.Username,
			Name:           user.Name,
			ProfilePhoto:   user.ProfilePhoto,
			Type:           notification.Type,
			Message:        notification.Message,
			Time:           notification.Time,
			Seen:           seen,
			ConversationId: notification.ConversationId,
		})
	}

	return notificationsData, nil
}

// GetConversation messages from notifications table
func GetConversation(db *gorm.DB, username, id string) ([]Conversation, error) {
	var conversation []Conversation
	var userId int64

	if err := db.
		Table("users").
		Select("id").
		Where("username = ?", username).
		Find(&userId).
		Error; err != nil {
		return nil, err
	}

	if err := db.
		Table("notifications").
		Select("id, sender, receiver, type, message, time, sender_id, reply_message, audio_message, reaction").
		Where("id = ? OR conversation_id = ?", id, id).
		Find(&conversation).
		Error; err != nil {
		return nil, err
	}

	c := addReceiver(conversation, username, userId)

	return c, nil
}

func addReceiver(conversation []Conversation, username string, userId int64) []Conversation {
	var newConversation []Conversation

	for _, c := range conversation {
		if len(c.Receiver) > 0 && len(c.Sender) > 0 {
			newConversation = append(newConversation, c)
			continue
		}

		if c.SenderId != userId {
			v := c
			v.Receiver = username

			newConversation = append(newConversation, v)
		} else {
			newConversation = append(newConversation, c)
		}
	}

	return newConversation
}

// GetUnseenNotifications get unseen notifications from notifications table
func GetUnseenNotifications(db *gorm.DB, username string) (*int64, error) {
	var unseenNotifications int64
	var receiverId int64

	if err := db.
		Table("users").
		Select("id").
		Where("username = ?", username).
		Find(&receiverId).
		Error; err != nil {
		return nil, err
	}

	if err := db.
		Table("notifications").
		Where("(receiver = ? OR receiver_id = ?) AND seen = 0", username, receiverId).
		Count(&unseenNotifications).
		Error; err != nil {
		return nil, err
	}

	return &unseenNotifications, nil
}

// UpdateSeenNotification update unseen notification in notifications table
func UpdateSeenNotification(db *gorm.DB, username, id string) error {
	var receiverId int64

	if err := db.
		Table("users").
		Select("id").
		Where("username = ?", username).
		Find(&receiverId).
		Error; err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		return tx.
			Table("notifications").
			Where("(receiver = ? OR receiver_id = ?) AND seen = 0 AND (conversation_id = ? OR id = ?)",
				username, receiverId, id, id).
			Update("seen", 1).
			Error
	})
}

// UpdateNotificationReaction update a notification reaction field in notifications table
func UpdateNotificationReaction(db *gorm.DB, username, id string, reaction *string) error {
	var notification Notification
	if err := db.
		Table("notifications").
		Where("id = ?", id).
		First(&notification).
		Error; err != nil {
		return err
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		return tx.
			Table("notifications").
			Where("id = ?", id).
			Update("reaction", reaction).
			Error
	}); err != nil {
		return err
	}

	var name string
	if err := db.
		Table("users").
		Select("name").
		Where("username = ?", username).
		Find(&name).
		Error; err != nil {
		return err
	}

	var tokens []string
	var err error
	if notification.SenderId > 0 {
		tokens, err = service.GetTokensByUserId(db, notification.SenderId)
	} else {
		tokens, err = service.GetTokensByUsername(db, notification.Sender)
	}
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		return nil
	}

	fcmNotification := service.FcmNotification{
		Title:   name + " reacted",
		Body:    "Your message received a reaction: " + *reaction,
		Sound:   "default",
		Devices: tokens,
	}

	return service.SendNotification(&fcmNotification)
}

func getNotificationUser(usersData []users.UserData, senderId, receiverId int64) users.UserData {
	for _, user := range usersData {
		if user.Id == senderId || user.Id == receiverId {
			return user
		}
	}

	return users.UserData{}
}

// getUsernamesFromNotifications get usernames from notifications array
func getUserIdsFromNotifications(notifications []Notification, userId int64) []int64 {
	var ids []int64

	for _, notification := range notifications {
		if !containsInt(ids, notification.SenderId) && notification.SenderId != userId {
			ids = append(ids, notification.SenderId)
		}
		if !containsInt(ids, notification.ReceiverId) && notification.ReceiverId != userId {
			ids = append(ids, notification.ReceiverId)
		}
	}

	return ids
}

// Helper function to check if string array contains value
func containsInt(s []int64, e int64) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}
