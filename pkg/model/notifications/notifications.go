package notifications

import (
	"fmt"

	"github.com/radekkrejcirik01/Koala-backend/pkg/model/users"
	"github.com/radekkrejcirik01/Koala-backend/pkg/service"
	"gorm.io/gorm"
)

const EmotionNotificationType = "emotion"
const StatusReplyNotificationType = "status_reply"
const MessageNotificationType = "message"

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
	ReplyMessage   *string
}

type NotificationData struct {
	Id             int64   `json:"id"`
	SenderId       int64   `json:"senderId"`
	Sender         string  `json:"sender"`
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	Message        string  `json:"message"`
	Time           int64   `json:"time"`
	Seen           int     `json:"seen"`
	ConversationId *int64  `json:"conversationId,omitempty"`
	Emotion        *string `json:"emotion,omitempty"`
	Expression     *string `json:"expression,omitempty"`
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
	Sender       string `json:"sender"`
	Receiver     string `json:"receiver"`
	Message      string `json:"message"`
	Time         int64  `json:"time"`
	ReplyMessage string `json:"replyMessage"`
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
	notification := Notification{
		SenderId:       t.SenderId,
		Sender:         username,
		ReceiverId:     t.ReceiverId,
		Receiver:       t.Receiver,
		Type:           MessageNotificationType,
		Message:        t.Message,
		ConversationId: &t.ConversationId,
		ReplyMessage:   t.ReplyMessage,
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

	fcmNotification := service.FcmNotification{
		Title:   t.Name,
		Body:    t.Message,
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
		Where(idCondition+`(receiver = ? OR receiver_id = ?)
			AND((id IN(
					SELECT
						MAX(id)
						FROM notifications
					WHERE
						(receiver = ? OR receiver_id = ?)
						AND TYPE = 'message'
					GROUP BY
						conversation_id))
				OR type IN ('emotion', 'status_reply'))`,
			username, receiverId, username, receiverId).
		Order("id DESC").
		Limit(20).
		Find(&notifications).
		Error; err != nil {
		return nil, err
	}

	usernames := getSendersFromNotifications(notifications)

	if err := db.
		Table("users").
		Where("username IN ?", usernames).
		Find(&usersData).
		Error; err != nil {
		return nil, err
	}

	conversationsIds := getConversationsIds(notifications)

	var emotionsData []EmotionData
	if len(conversationsIds) > 0 {
		if err := db.
			Table("notifications").
			Select("id, message").
			Where("id IN ?", conversationsIds).
			Find(&emotionsData).
			Error; err != nil {
			return nil, err
		}
	}

	notificationsIds := getNotificationsIds(notifications)

	var expressionsData []ExpressionData
	if err := db.
		Table("notifications").
		Select("id, reply_message").
		Where("id IN ? AND type = 'status_reply'", notificationsIds).
		Find(&expressionsData).
		Error; err != nil {
		return nil, err
	}

	var notificationsData []NotificationData
	for _, notification := range notifications {
		notificationsData = append(notificationsData, NotificationData{
			Id:             int64(notification.Id),
			SenderId:       notification.SenderId,
			Sender:         notification.Sender,
			Name:           getName(usersData, notification.Sender),
			Type:           notification.Type,
			Message:        notification.Message,
			Time:           notification.Time,
			Seen:           notification.Seen,
			ConversationId: notification.ConversationId,
			Emotion:        getEmotionMessage(emotionsData, notification.ConversationId),
			Expression:     getExpression(expressionsData, int64(notification.Id)),
		})
	}

	return notificationsData, nil
}

// GetFilteredNotifications get notifications by user id from notifications table
func GetFilteredNotifications(db *gorm.DB, username, userId, lastId string) ([]NotificationData, error) {
	var notifications []Notification
	var usersData users.UserData

	var idCondition string
	if lastId != "" {
		idCondition = fmt.Sprintf("id < %s AND ", lastId)
	}

	if err := db.
		Table("users").
		Where("id = ?", userId).
		Find(&usersData).
		Error; err != nil {
		return nil, nil
	}

	if err := db.
		Table("notifications").
		Where(idCondition+`receiver = ? AND sender = ?
		AND((id IN(
				SELECT
					MAX(id)
					FROM notifications
				WHERE
					receiver = ?
					AND sender = ?
					AND TYPE = 'message'
				GROUP BY
					conversation_id))
			OR TYPE = 'emotion')`,
			username, usersData.Username, username, usersData.Username).
		Order("id DESC").
		Limit(20).
		Find(&notifications).
		Error; err != nil {
		return nil, err
	}

	conversationsIds := getConversationsIds(notifications)

	var emotionsData []EmotionData
	if len(conversationsIds) > 0 {
		if err := db.
			Table("notifications").
			Select("id, message").
			Where("id IN ?", conversationsIds).
			Find(&emotionsData).
			Error; err != nil {
			return nil, err
		}
	}

	var notificationsData []NotificationData
	for _, notification := range notifications {
		notificationsData = append(notificationsData, NotificationData{
			Id:             int64(notification.Id),
			SenderId:       notification.SenderId,
			Sender:         notification.Sender,
			Name:           usersData.Name,
			Type:           notification.Type,
			Message:        notification.Message,
			Time:           notification.Time,
			Seen:           notification.Seen,
			ConversationId: notification.ConversationId,
			Emotion:        getEmotionMessage(emotionsData, notification.ConversationId),
		})
	}

	return notificationsData, nil
}

// GetConversation messages from notifications table
func GetConversation(db *gorm.DB, username, id string) ([]Conversation, error) {
	var conversation []Conversation

	if err := db.
		Table("notifications").
		Select("id, sender, receiver, message, time, reply_message").
		Where("id = ? OR conversation_id = ?", id, id).
		Find(&conversation).
		Error; err != nil {
		return nil, err
	}

	return conversation, nil
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
	var notifications []Notification
	var usersData []users.UserData

	var idCondition string
	if lastId != "" {
		idCondition = fmt.Sprintf("id < %s AND ", lastId)
	}

	if err := db.
		Table("notifications").
		Where(idCondition+"sender = ? AND type = 'emotion'", username).
		Order("id DESC").
		Limit(20).
		Find(&notifications).
		Error; err != nil {
		return nil, err
	}

	n := groupNotifications(notifications)

	usernames := getReceiversFromNotifications(notifications)

	if err := db.
		Table("users").
		Where("username IN ?", usernames).
		Find(&usersData).
		Error; err != nil {
		return nil, err
	}

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

// GetTrack gets track of sent notifications from notifications table
func GetTrack(db *gorm.DB, username string, lastId string) ([]TrackData, error) {
	var notifications []Notification
	var usersData []users.UserData

	var idCondition string
	if lastId != "" {
		idCondition = fmt.Sprintf("id < %s AND ", lastId)
	}

	if err := db.
		Table("notifications").
		Where(idCondition+"sender = ? AND type = 'emotion'", username).
		Order("id DESC").
		Limit(20).
		Find(&notifications).
		Error; err != nil {
		return nil, err
	}

	n := groupNotifications(notifications)

	usernames := getReceiversFromNotifications(notifications)

	if err := db.
		Table("users").
		Where("username IN ?", usernames).
		Find(&usersData).
		Error; err != nil {
		return nil, err
	}

	var track []TrackData
	for _, notification := range n {
		track = append(track, TrackData{
			Id:             int(notification.Id),
			ReceiversNames: getNamesFromNotifications(notifications, usersData, notification),
			Message:        notification.Message,
			Time:           notification.Time,
		})
	}

	return track, nil
}

// Helper function to get names from notifications
func getNamesFromNotifications(notifications []Notification, users []users.UserData, notification Notification) []string {
	var receivers []string
	var names []string

	for _, n := range notifications {
		if n.Message == notification.Message && n.Time == notification.Time {
			receivers = append(receivers, n.Receiver)
		}
	}

	for _, user := range users {
		if contains(receivers, user.Username) {
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

// Helper function to get name from users data
func getName(usersData []users.UserData, username string) string {
	for _, user := range usersData {
		if user.Username == username {
			return user.Name
		}
	}

	return ""
}

// getEmotionMessage helper function to get emotion message by conversation id
func getEmotionMessage(emotionsData []EmotionData, conversationId *int64) *string {
	if conversationId != nil {
		for _, v := range emotionsData {
			if v.Id == *conversationId {
				return &v.Message
			}
		}
	}
	return nil
}

// getExpression helper function to get expression message by notifications id
func getExpression(expressionsData []ExpressionData, notificationId int64) *string {
	for _, v := range expressionsData {
		if v.Id == notificationId {
			return &v.ReplyMessage
		}
	}
	return nil
}

// getConversationsIds helper function to get conversation ids from notifications
func getConversationsIds(notifications []Notification) []int64 {
	var ids []int64

	for _, v := range notifications {
		if v.ConversationId != nil {
			ids = append(ids, *v.ConversationId)
		}
	}

	return ids
}

// getNotificationsIds helper function to get notifications ids from notifications
func getNotificationsIds(notifications []Notification) []int64 {
	var ids []int64

	for _, v := range notifications {
		ids = append(ids, int64(v.Id))
	}

	return ids
}

// Helper function to get senders from notifications
func getSendersFromNotifications(notifications []Notification) []string {
	var senders []string

	for _, notification := range notifications {
		if !contains(senders, notification.Sender) {
			senders = append(senders, notification.Sender)
		}
	}

	return senders
}

// Helper function to get receivers from notifications
func getReceiversFromNotifications(notifications []Notification) []string {
	var receivers []string

	for _, notification := range notifications {
		if !contains(receivers, notification.Receiver) {
			receivers = append(receivers, notification.Receiver)
		}
	}

	return receivers
}

// Helper function to check if string array contains value
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}
