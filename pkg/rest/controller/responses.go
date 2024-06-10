package controller

import (
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/emotions"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/expressions"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/invites"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/notifications"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/replies"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/users"
)

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type AuthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Token   string `json:"token,omitempty"`
}

type UserResponse struct {
	Status  string         `json:"status"`
	Message string         `json:"message,omitempty"`
	Data    users.UserData `json:"data,omitempty"`
}

type LastOnlineResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Time    int64  `json:"time"`
}

type FriendsResponse struct {
	Status  string           `json:"status"`
	Message string           `json:"message,omitempty"`
	Data    []users.UserData `json:"data,omitempty"`
}

type InvitesResponse struct {
	Status  string               `json:"status"`
	Message string               `json:"message,omitempty"`
	Data    []invites.InviteData `json:"data,omitempty"`
}

type NotificationsResponse struct {
	Status  string                           `json:"status"`
	Message string                           `json:"message,omitempty"`
	Data    []notifications.NotificationData `json:"data,omitempty"`
}

type ConversationResponse struct {
	Status  string                       `json:"status"`
	Message string                       `json:"message,omitempty"`
	Data    []notifications.Conversation `json:"data,omitempty"`
}

type UnseenNotificationsResponse struct {
	Status              string `json:"status"`
	Message             string `json:"message,omitempty"`
	UnseenNotifications int64  `json:"unseenNotifications,omitempty"`
}

type HistoryResponse struct {
	Status  string                      `json:"status"`
	Message string                      `json:"message,omitempty"`
	Data    []notifications.HistoryData `json:"data,omitempty"`
}

type EmotionsResponse struct {
	Status  string                  `json:"status"`
	Message string                  `json:"message,omitempty"`
	Data    []emotions.EmotionsData `json:"data,omitempty"`
	Removed []int                   `json:"removed,omitempty"`
}

type ExpressionsResponse struct {
	Status     string                        `json:"status"`
	Message    string                        `json:"message,omitempty"`
	Data       []expressions.ExpressionsData `json:"data,omitempty"`
	Expression string                        `json:"expression,omitempty"`
}

type RecordingResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Url     string `json:"url,omitempty"`
}

type RepliesResponse struct {
	Status  string              `json:"status"`
	Message string              `json:"message,omitempty"`
	Data    []replies.ReplyData `json:"data,omitempty"`
}
