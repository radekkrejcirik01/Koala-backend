package notifications

import (
	"fmt"

	"github.com/radekkrejcirik01/Koala-backend/pkg/model/users"
	"github.com/radekkrejcirik01/Koala-backend/pkg/service"
	"gorm.io/gorm"
)

const EmotionsNotificationType = "emotion"
const SupportNotificationType = "support"

type Notification struct {
	Id       uint `gorm:"primary_key;auto_increment;not_null"`
	Sender   string
	Receiver string
	Type     string
	Message  string
	Time     int64 `gorm:"autoCreateTime"`
}

func (Notification) TableName() string {
	return "notifications"
}

type EmotionNotification struct {
	Receivers []string
	Name      string
	Message   string
}

type SupportNotification struct {
	Receiver string
	Name     string
	Message  string
}

type NotificationData struct {
	Id      int    `json:"id"`
	Sender  string `json:"sender"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Message string `json:"message"`
	Time    int64  `json:"time"`
}

type HistoryData struct {
	Id             int      `json:"id"`
	ReceiversNames []string `json:"receiversNames"`
	Message        string   `json:"message"`
	Time           int64    `json:"time"`
}

// SendEmotionNotification sends emotion notification
func SendEmotionNotification(db *gorm.DB, t *EmotionNotification, username string) error {
	var n []Notification

	for _, receiver := range t.Receivers {
		n = append(n, Notification{
			Sender:   username,
			Receiver: receiver,
			Type:     EmotionsNotificationType,
			Message:  t.Message,
		})
	}

	if err := db.Table("notifications").Create(&n).Error; err != nil {
		return err
	}

	tokens, err := service.GetTokensByUsernames(db, t.Receivers)
	if err != nil {
		return err
	}

	fcmNotification := service.FcmNotification{
		Title:   t.Name + " is sharing",
		Body:    t.Message,
		Sound:   "default",
		Devices: tokens,
	}

	return service.SendNotification(&fcmNotification)
}

// SendSupportNotification sends support notification
func SendSupportNotification(db *gorm.DB, t *SupportNotification, username string) error {
	notification := Notification{
		Sender:   username,
		Receiver: t.Receiver,
		Type:     SupportNotificationType,
		Message:  t.Message,
	}

	if err := db.Table("notifications").Create(&notification).Error; err != nil {
		return err
	}

	tokens, err := service.GetTokensByUsername(db, t.Receiver)
	if err != nil {
		return err
	}

	fcmNotification := service.FcmNotification{
		Title:   t.Name + " is sending support ❤️",
		Body:    `"` + t.Message + `""`,
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

	if err := db.
		Table("notifications").
		Where(idCondition+"receiver = ?", username).
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

	var notificationsData []NotificationData
	for _, notification := range notifications {
		notificationsData = append(notificationsData, NotificationData{
			Id:      int(notification.Id),
			Sender:  notification.Sender,
			Name:    getName(usersData, notification.Sender),
			Type:    notification.Type,
			Message: notification.Message,
			Time:    notification.Time,
		})
	}

	return notificationsData, nil
}

// GetHistory gets notifications history from notifications table
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

// Helper function to check check if notifications array contains notification
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
