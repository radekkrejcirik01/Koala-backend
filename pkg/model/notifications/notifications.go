package notifications

import (
	"fmt"

	"github.com/radekkrejcirik01/Koala-backend/pkg/model/recordings"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/users"
	"github.com/radekkrejcirik01/Koala-backend/pkg/service"
	"gorm.io/gorm"
)

const EmotionNotificationType = "emotion"
const StatusReplyNotificationType = "status_reply"
const MessageNotificationType = "message"
const AudioMessageNotificationType = "audio"

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
}

func (Notification) TableName() string {
	return "notifications"
}

type EmotionNotification struct {
	SenderId     int64
	ReceiversIds []int64
	Receivers    []string
	Name         string
	Message      string
}

type StatusReplyNotification struct {
	SenderId        int64
	ReceiverId      int64
	Name            string
	Message         string
	ReplyExpression string
}

type MessageNotification struct {
	SenderId       int64
	ReceiverId     int64
	Receiver       string
	Name           string
	Message        string
	ConversationId int64
	ReplyMessage   string
	AudioBuffer    string
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
}

type HistoryData struct {
	Id             int      `json:"id"`
	ReceiversNames []string `json:"receiversNames"`
	Message        string   `json:"message"`
	Time           int64    `json:"time"`
}

type TrackData struct {
	Id             int      `json:"id"`
	ReceiversNames []string `json:"receiversNames"`
	Message        string   `json:"message"`
	Time           int64    `json:"time"`
}

func getReceiverId(users []users.UserData, username string) int64 {
	for _, v := range users {
		if v.Username == username {
			return v.Id
		}
	}
	return 0
}

func getReceiver(users []users.UserData, userId int64) string {
	for _, v := range users {
		if v.Id == userId {
			return v.Username
		}
	}
	return ""
}

// SendEmotionNotification sends emotion notification
func SendEmotionNotification(db *gorm.DB, t *EmotionNotification, username string) error {
	var n []Notification
	var users []users.UserData

	if err := db.
		Table("users").
		Select("id, username").
		Distinct().
		Where("id IN ? OR username IN ?", t.ReceiversIds, t.Receivers).
		Find(&users).
		Error; err != nil {
		return err
	}

	if len(t.ReceiversIds) > 0 {
		for _, receiverId := range t.ReceiversIds {
			n = append(n, Notification{
				SenderId:   t.SenderId,
				Sender:     username,
				ReceiverId: receiverId,
				Receiver:   getReceiver(users, receiverId),
				Type:       EmotionNotificationType,
				Message:    t.Message,
			})
		}
	} else {
		for _, receiver := range t.Receivers {
			n = append(n, Notification{
				SenderId:   t.SenderId,
				Sender:     username,
				ReceiverId: getReceiverId(users, receiver),
				Receiver:   receiver,
				Type:       EmotionNotificationType,
				Message:    t.Message,
			})
		}
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		return tx.Table("notifications").Create(&n).Error
	})
	if err != nil {
		return err
	}

	var tokens []string
	if len(t.ReceiversIds) > 0 {
		tokens, err = service.GetTokensByUserIds(db, t.ReceiversIds)
		if err != nil {
			return err
		}
	} else {
		tokens, err = service.GetTokensByUsernames(db, t.Receivers)
		if err != nil {
			return err
		}
	}

	fcmNotification := service.FcmNotification{
		Title:   t.Name + " is sharing",
		Body:    t.Message,
		Sound:   "default",
		Devices: tokens,
	}

	return service.SendNotification(&fcmNotification)
}

// SendStatusReplyNotification sends status reply message notification
func SendStatusReplyNotification(db *gorm.DB, t *StatusReplyNotification, username string) error {
	notification := Notification{
		SenderId:     t.SenderId,
		Sender:       username,
		ReceiverId:   t.ReceiverId,
		Type:         StatusReplyNotificationType,
		Message:      t.Message,
		ReplyMessage: &t.ReplyExpression,
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		return tx.Table("notifications").Create(&notification).Error
	})
	if err != nil {
		return err
	}

	tokens, err := service.GetTokensByUserId(db, t.ReceiverId)
	if err != nil {
		return err
	}

	fcmNotification := service.FcmNotification{
		Body:    t.Name + " is replying to your status: " + t.Message,
		Sound:   "default",
		Devices: tokens,
	}

	return service.SendNotification(&fcmNotification)
}

// SendMessageNotification sends message notification
func SendMessageNotification(db *gorm.DB, t *MessageNotification, username string) error {
	messageType := MessageNotificationType
	var audioMessageUrl string

	if isAudioMessage(t.AudioBuffer) {
		var err error
		messageType = AudioMessageNotificationType

		// Ensure message is emptied when sending voice message
		t.Message = ""

		audioMessageUrl, err = recordings.UploadRecording(t.AudioBuffer, t.SenderId)
		if err != nil {
			return err
		}
	}

	notification := Notification{
		SenderId:       t.SenderId,
		Sender:         username,
		ReceiverId:     t.ReceiverId,
		Receiver:       t.Receiver,
		Type:           messageType,
		Message:        t.Message,
		ConversationId: &t.ConversationId,
		ReplyMessage:   &t.ReplyMessage,
		AudioMessage:   &audioMessageUrl,
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		return tx.Table("notifications").Create(&notification).Error
	})
	if err != nil {
		return err
	}

	var tokens []string
	if t.ReceiverId > 0 {
		tokens, err = service.GetTokensByUserId(db, t.ReceiverId)
		if err != nil {
			return err
		}
	} else {
		tokens, err = service.GetTokensByUsername(db, t.Receiver)
		if err != nil {
			return err
		}
	}

	body := t.Message
	if isAudioMessage(audioMessageUrl) {
		body = "🎤 Voice message"
	}

	fcmNotification := service.FcmNotification{
		Title:   t.Name,
		Body:    body,
		Sound:   "default",
		Devices: tokens,
	}

	return service.SendNotification(&fcmNotification)
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

// GetFriendNotifications gets notifications from notifications table by friend
func GetFriendNotifications(db *gorm.DB, username string, friendId, lastId string) ([]NotificationData, error) {
	var notifications []Notification
	var usersData users.UserData

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
			((sender_id = ? AND (type = 'message' OR type = 'audio')) OR receiver_id = ?) AND (sender_id = ? OR receiver_id = ?)
		GROUP BY
			conversation_id)`,
			userId, userId, userId, userId, friendId, friendId).
		Order("id DESC").
		Limit(20).
		Find(&notifications).
		Error; err != nil {
		return nil, err
	}

	if err := db.
		Table("users").
		Where("id = ?", friendId).
		Find(&usersData).
		Error; err != nil {
		return nil, err
	}

	var notificationsData []NotificationData
	for _, notification := range notifications {
		seen := notification.Seen
		if notification.SenderId == userId {
			seen = 1
		}

		notificationsData = append(notificationsData, NotificationData{
			Id:             int64(notification.Id),
			SenderId:       usersData.Id,
			Sender:         usersData.Username,
			Name:           usersData.Name,
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
		Select("id, sender, receiver, type, message, time, sender_id, reply_message, audio_message").
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

// GetHistory get history of sahred emotions from notifications table
func GetHistory(db *gorm.DB, username string, lastId string) ([]HistoryData, error) {
	var userId int64
	var notifications []Notification
	var usersData []users.UserData

	var idCondition string
	if lastId != "" {
		idCondition = fmt.Sprintf("id < %s AND ", lastId)
	}

	if err := db.
		Table("users").
		Select("id").
		Where("username = ?", username).
		Find(&userId).
		Error; err != nil {
		return nil, nil
	}

	if err := db.
		Table("notifications").
		Where(idCondition+"sender_id = ? AND type IN ('emotion', 'direct_emotion', 'check_on')", userId).
		Order("id DESC").
		Limit(20).
		Find(&notifications).
		Error; err != nil {
		return nil, err
	}

	ids := getReceiversIdsFromNotifications(notifications)

	if err := db.
		Table("users").
		Where("id IN ?", ids).
		Find(&usersData).
		Error; err != nil {
		return nil, err
	}

	n := groupNotifications(notifications)

	var history []HistoryData
	for _, notification := range n {
		history = append(history, HistoryData{
			Id:             int(notification.Id),
			ReceiversNames: getNamesFromNotifications(notifications, usersData, notification),
			Message:        notification.Message,
			Time:           notification.Time,
		})
	}

	return history, nil
}

// GetUserHistory get history of sahred emotions to friend from notifications table
func GetUserHistory(db *gorm.DB, username, receiverId, lastId string) ([]HistoryData, error) {
	var userId int64
	var notifications []Notification

	var idCondition string
	if lastId != "" {
		idCondition = fmt.Sprintf("id < %s AND ", lastId)
	}

	if err := db.
		Table("users").
		Select("id").
		Where("username = ?", username).
		Find(&userId).
		Error; err != nil {
		return nil, nil
	}

	if err := db.
		Table("notifications").
		Where(idCondition+"sender_id = ? AND receiver_id = ? AND type IN ('emotion', 'direct_emotion', 'check_on')", userId, receiverId).
		Order("id DESC").
		Limit(20).
		Find(&notifications).
		Error; err != nil {
		return nil, err
	}

	var history []HistoryData
	for _, notification := range notifications {
		history = append(history, HistoryData{
			Id:      int(notification.Id),
			Message: notification.Message,
			Time:    notification.Time,
		})
	}

	return history, nil
}

// Helper function to get names from notifications
func getNamesFromNotifications(notifications []Notification, users []users.UserData, notification Notification) []string {
	var receiversIds []int64
	var names []string

	for _, n := range notifications {
		if n.Message == notification.Message && n.Time == notification.Time {
			receiversIds = append(receiversIds, n.ReceiverId)
		}
	}

	for _, user := range users {
		if containsInt(receiversIds, user.Id) {
			names = append(names, user.Name)
		}
	}

	return names
}

// Helper function to group same notifications with different receivers
func groupNotifications(notifications []Notification) []Notification {
	var n []Notification

	for _, notification := range notifications {
		if !containsNotification(n, notification) {
			n = append(n, notification)
		}
	}

	return n
}

// Helper function to check if notifications array contains notification
func containsNotification(notifications []Notification, notification Notification) bool {
	for _, n := range notifications {
		if n.Message == notification.Message && n.Time == notification.Time {
			return true
		}
	}
	return false
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

// Helper function to get receivers from notifications
func getReceiversIdsFromNotifications(notifications []Notification) []int64 {
	var receiversIds []int64

	for _, notification := range notifications {
		if !containsInt(receiversIds, notification.ReceiverId) {
			receiversIds = append(receiversIds, notification.ReceiverId)
		}
	}

	return receiversIds
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

// Check if audio message has lenght
func isAudioMessage(message string) bool {
	return len(message) > 0
}
