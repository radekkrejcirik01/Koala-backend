package controller

import "github.com/radekkrejcirik01/Koala-backend/pkg/model/users"

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
