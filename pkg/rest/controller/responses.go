package controller

import (
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/emotions"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/notifications"
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
	Status   string                  `json:"status"`
	Message  string                  `json:"message,omitempty"`
	Data     users.UserData          `json:"data,omitempty"`
	Emotions []emotions.EmotionsData `json:"emotions,omitempty"`
}

type FriendsResponse struct {
	Status  string           `json:"status"`
	Message string           `json:"message,omitempty"`
	Data    []users.UserData `json:"data,omitempty"`
}

type NotificationsResponse struct {
	Status  string                           `json:"status"`
	Message string                           `json:"message,omitempty"`
	Data    []notifications.NotificationData `json:"data,omitempty"`
}

type UnseenNotificationsResponse struct {
	Status              string `json:"status"`
	Message             string `json:"message,omitempty"`
	UnseenNotifications int64  `json:"unseenNotifications,omitempty"`
}

type TrackResponse struct {
	Status  string                    `json:"status"`
	Message string                    `json:"message,omitempty"`
	Data    []notifications.TrackData `json:"data,omitempty"`
}

type EmotionsResponse struct {
	Status  string                  `json:"status"`
	Message string                  `json:"message,omitempty"`
	Data    []emotions.EmotionsData `json:"data,omitempty"`
}
