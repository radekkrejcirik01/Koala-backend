package controller

import (
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/chats"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/emotions"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/invites"
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

type EmotionsResponse struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message,omitempty"`
	Data    []emotions.EmotionData `json:"data,omitempty"`
}

type ChatsResponse struct {
	Status  string           `json:"status"`
	Message string           `json:"message,omitempty"`
	Data    []chats.ChatData `json:"data,omitempty"`
}

type UnreadChatsResponse struct {
	Status      string `json:"status"`
	Message     string `json:"message,omitempty"`
	UnreadChats int64  `json:"unreadChats"`
}

type ChatMessagesResponse struct {
	Status  string              `json:"status"`
	Message string              `json:"message,omitempty"`
	Data    []chats.MessageData `json:"data,omitempty"`
}

type ChatMessageResponse struct {
	Status  string            `json:"status"`
	Message string            `json:"message,omitempty"`
	Data    chats.MessageData `json:"data"`
}

type UploadProfilePhotoResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Data    string `json:"data,omitempty"`
}
