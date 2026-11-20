package connections

import (
	"encoding/json"
	"log"
	"time"

	"github.com/radekkrejcirik01/Koala-backend/pkg/service"
	"gorm.io/gorm"
)

// Events sent to the app through the WebSocket API
const (
	MessageNewEvent      = "message.new"
	MessageDeletedEvent  = "message.deleted"
	MessageReactionEvent = "message.reaction"
	ChatReadEvent        = "chat.read"
	TypingEvent          = "typing"
	PresenceEvent        = "presence"
)

// maxAge is the longest time API Gateway keeps a connection open,
// older connections are stale when their $disconnect didn't arrive
const maxAge = 2 * time.Hour

type Connection struct {
	ConnectionId string `gorm:"primary_key;size:128"`
	UserId       int64  `gorm:"index"`
	ConnectedAt  int64  `gorm:"autoCreateTime"`
}

func (Connection) TableName() string {
	return "connections"
}

type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Delivery is an event for all connections of the user
type Delivery struct {
	UserId int64
	Event  Event
}

// Save saves user's new connection and removes user's stale ones
func Save(db *gorm.DB, connectionId string, userId int64) error {
	if err := db.
		Table("connections").
		Where("user_id = ? AND connected_at < ?", userId, staleBefore()).
		Delete(&Connection{}).
		Error; err != nil {
		return err
	}

	return db.Table("connections").Create(&Connection{
		ConnectionId: connectionId,
		UserId:       userId,
	}).Error
}

// Delete deletes the connection and returns its user, user id is 0 when the connection doesn't exist
func Delete(db *gorm.DB, connectionId string) (int64, error) {
	var connection Connection
	if err := db.
		Table("connections").
		Where("connection_id = ?", connectionId).
		Limit(1).
		Find(&connection).
		Error; err != nil {
		return 0, err
	}

	if connection.UserId == 0 {
		return 0, nil
	}

	return connection.UserId, db.
		Table("connections").
		Where("connection_id = ?", connectionId).
		Delete(&Connection{}).
		Error
}

// DeleteUserConnections deletes all connections of the user
func DeleteUserConnections(tx *gorm.DB, userId int64) error {
	return tx.
		Table("connections").
		Where("user_id = ?", userId).
		Delete(&Connection{}).
		Error
}

// CountUserConnections counts user's open connections
func CountUserConnections(db *gorm.DB, userId int64) (int64, error) {
	var count int64
	err := db.
		Table("connections").
		Where("user_id = ? AND connected_at >= ?", userId, staleBefore()).
		Count(&count).
		Error

	return count, err
}

// GetOnlineUserIds gets which of the users have an open connection
func GetOnlineUserIds(db *gorm.DB, userIds []int64) ([]int64, error) {
	if len(userIds) == 0 {
		return []int64{}, nil
	}

	ids := []int64{}
	err := db.
		Table("connections").
		Select("user_id").
		Distinct("user_id").
		Where("user_id IN ? AND connected_at >= ?", userIds, staleBefore()).
		Find(&ids).
		Error

	return ids, err
}

// Send sends the events to all connections of their users, failures are only logged
// so they never fail the request, gone connections are deleted
func Send(db *gorm.DB, deliveries []Delivery) {
	if len(deliveries) == 0 {
		return
	}

	userEvents := make(map[int64][][]byte)
	userIds := make([]int64, 0, len(deliveries))
	for _, delivery := range deliveries {
		data, err := json.Marshal(delivery.Event)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		if _, ok := userEvents[delivery.UserId]; !ok {
			userIds = append(userIds, delivery.UserId)
		}
		userEvents[delivery.UserId] = append(userEvents[delivery.UserId], data)
	}

	var connections []Connection
	if err := db.
		Table("connections").
		Where("user_id IN ? AND connected_at >= ?", userIds, staleBefore()).
		Find(&connections).
		Error; err != nil {
		log.Println(err.Error())
		return
	}

	var messages []service.SocketMessage
	for _, connection := range connections {
		for _, data := range userEvents[connection.UserId] {
			messages = append(messages, service.SocketMessage{
				ConnectionId: connection.ConnectionId,
				Data:         data,
			})
		}
	}

	gone := service.SendToConnections(messages)
	if len(gone) == 0 {
		return
	}

	if err := db.
		Table("connections").
		Where("connection_id IN ?", gone).
		Delete(&Connection{}).
		Error; err != nil {
		log.Println(err.Error())
	}
}

func staleBefore() int64 {
	return time.Now().Add(-maxAge).Unix()
}
