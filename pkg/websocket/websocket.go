package websocket

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
	"github.com/radekkrejcirik01/Koala-backend/pkg/middleware"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/chats"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/connections"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/invites"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/users"
	"gorm.io/gorm"
)

var errUnauthorized = errors.New("unauthorized")

type TypingRequest struct {
	FriendId int64
	Typing   bool
}

type PresenceEventData struct {
	UserId     int64 `json:"userId"`
	Online     bool  `json:"online"`
	LastOnline int64 `json:"lastOnline"`
}

// Handle handles the WebSocket API routes, $default takes the app's keep-alive messages
func Handle(request events.APIGatewayWebsocketProxyRequest) events.APIGatewayProxyResponse {
	connectionId := request.RequestContext.ConnectionID

	var err error
	switch request.RequestContext.RouteKey {
	case "$connect":
		err = connect(database.DB, connectionId, request)
	case "$disconnect":
		err = disconnect(database.DB, connectionId)
	case "typing":
		err = typing(database.DB, connectionId, request.Body)
	}

	if err != nil {
		log.Println(err.Error())
		status := http.StatusInternalServerError
		if errors.Is(err, errUnauthorized) {
			status = http.StatusUnauthorized
		}
		return events.APIGatewayProxyResponse{StatusCode: status}
	}

	return events.APIGatewayProxyResponse{StatusCode: http.StatusOK}
}

// connect authorizes the connection with the JWT token from Authorization header or token query parameter
func connect(db *gorm.DB, connectionId string, request events.APIGatewayWebsocketProxyRequest) error {
	token := strings.TrimPrefix(header(request.Headers, "Authorization"), "Bearer ")
	if token == "" {
		token = request.QueryStringParameters["token"]
	}
	if token == "" {
		return errUnauthorized
	}

	username, err := middleware.AuthorizeToken(token)
	if err != nil {
		return errUnauthorized
	}

	user, err := users.GetUser(db, username)
	if err != nil {
		return err
	}
	if user.Id == 0 {
		return errUnauthorized
	}

	if err := connections.Save(db, connectionId, user.Id); err != nil {
		return err
	}

	count, err := connections.CountUserConnections(db, user.Id)
	if err != nil {
		return err
	}

	// Friends already know the user is online from another device
	if count > 1 {
		return nil
	}

	return sendPresence(db, user, true)
}

// disconnect removes the connection, the user is offline when it was user's last connection
func disconnect(db *gorm.DB, connectionId string) error {
	userId, err := connections.Delete(db, connectionId)
	if err != nil || userId == 0 {
		return err
	}

	count, err := connections.CountUserConnections(db, userId)
	if err != nil || count > 0 {
		return err
	}

	user, err := getUserById(db, userId)
	if err != nil || user.Id == 0 {
		return err
	}

	if err := users.UpdateLastOnline(db, user.Username); err != nil {
		return err
	}

	return sendPresence(db, user, false)
}

func typing(db *gorm.DB, connectionId, body string) error {
	t := &TypingRequest{}
	if err := json.Unmarshal([]byte(body), t); err != nil {
		return err
	}

	user, err := getConnectionUser(db, connectionId)
	if err != nil || user.Id == 0 {
		return err
	}

	return chats.SendTyping(db, user, t.FriendId, t.Typing)
}

// sendPresence sends user's online state to user's friends
func sendPresence(db *gorm.DB, user users.UserData, online bool) error {
	friends, err := invites.GetFriends(db, user.Username)
	if err != nil {
		return err
	}

	event := connections.Event{
		Type: connections.PresenceEvent,
		Data: PresenceEventData{
			UserId:     user.Id,
			Online:     online,
			LastOnline: time.Now().Unix(),
		},
	}

	deliveries := make([]connections.Delivery, 0, len(friends))
	for _, friend := range friends {
		if friend.Id != 0 {
			deliveries = append(deliveries, connections.Delivery{UserId: friend.Id, Event: event})
		}
	}
	connections.Send(db, deliveries)

	return nil
}

func getUserById(db *gorm.DB, id int64) (users.UserData, error) {
	var user users.UserData
	err := db.
		Table("users").
		Select("id, username, name").
		Where("id = ?", id).
		Limit(1).
		Find(&user).
		Error

	return user, err
}

// getConnectionUser gets the user of the connection, user Id is 0 when the connection doesn't exist
func getConnectionUser(db *gorm.DB, connectionId string) (users.UserData, error) {
	var user users.UserData
	err := db.
		Table("users").
		Joins("JOIN connections ON connections.user_id = users.id").
		Select("users.id, users.username, users.name").
		Where("connections.connection_id = ?", connectionId).
		Limit(1).
		Find(&user).
		Error

	return user, err
}

// header gets the header case-insensitively, API Gateway passes them as the app sent them
func header(headers map[string]string, name string) string {
	for key, value := range headers {
		if strings.EqualFold(key, name) {
			return value
		}
	}

	return ""
}
