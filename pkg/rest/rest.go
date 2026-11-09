package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/radekkrejcirik01/Koala-backend/pkg/rest/controller"
)

// Create new REST API server
func Create() *fiber.App {
	app := fiber.New()

	app.Get("/user", controller.GetUser)
	app.Get("/friends", controller.GetFriends)
	app.Get("/invites", controller.GetInvites)
	app.Get("/notifications/:lastId?", controller.GetNotifications)
	app.Get("/conversation/:id", controller.GetConversation)
	app.Get("/unseen-notifications", controller.GetUnseenNotifications)
	app.Get("/last-online/:id", controller.GetLastOnline)
	app.Get("/emotions", controller.GetEmotions)

	app.Post("/user", controller.CreateUser)
	app.Post("/login", controller.LoginUser)
	app.Post("/username", controller.CheckUsername)
	app.Post("/invite", controller.SendInvite)
	app.Post("/device", controller.SaveDevice)
	app.Post("/emotion-message/:type?", controller.SendEmotionMessage)
	app.Post("/message", controller.SendMessage)
	app.Post("/password-reset", controller.SendPasswordResetEmail)
	app.Post("/support", controller.SendSupport)
	app.Post("/feedback", controller.SendFeedback)
	app.Post("/profile-photo", controller.UploadProfilePhoto)

	app.Put("/invite", controller.AcceptInvite)
	app.Put("/notification/:id", controller.UpdateSeenNotification)
	app.Put("/notification-reaction/:id", controller.UpdateNotificationReaction)
	app.Put("/notification-support/:id", controller.UpdateNotificationSupport)
	app.Put("/user-name", controller.ChangeName)
	app.Put("/user-password", controller.ChangePassword)
	app.Put("/last-online", controller.UpdateLastOnline)

	app.Delete("/account", controller.DeleteAccount)
	app.Delete("/device", controller.DeleteDevice)
	app.Delete("/friend/:id", controller.RemoveFriend)
	app.Delete("/invite/:id", controller.RemoveInvite)
	app.Delete("/message/:id", controller.DeleteMessage)

	app.Get("/chats", controller.GetChats)
	app.Get("/chats/unread", controller.GetUnreadChats)
	app.Get("/chats/:friendId/messages", controller.GetChatMessages)
	app.Post("/chats/emotion", controller.SendChatEmotion)
	app.Post("/chats/:friendId/messages", controller.SendChatMessage)
	app.Post("/chats/messages/:id/support", controller.SendChatSupport)
	app.Put("/chats/:friendId/read", controller.ReadChat)
	app.Put("/chats/messages/:id/reaction", controller.UpdateChatReaction)
	app.Delete("/chats/messages/:id", controller.DeleteChatMessage)

	return app
}
