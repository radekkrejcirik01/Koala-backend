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
	app.Get("/history/:lastId?", controller.GetHistory)
	app.Get("/user-history/:receiverId/:lastId?", controller.GetUserHistory)
	app.Get("/emotions", controller.GetEmotions)
	app.Get("/expressions", controller.GetExpressions)

	app.Post("/user", controller.CreateUser)
	app.Post("/login", controller.LoginUser)
	app.Post("/invite", controller.SendInvite)
	app.Post("/device", controller.SaveDevice)
	app.Post("/emotion-notification", controller.SendEmotionNotification)
	app.Post("/status-reply-notification", controller.SendStatusReplyNotification)
	app.Post("/message-notification", controller.SendMessageNotification)
	app.Post("/emotion", controller.AddEmotion)
	app.Post("/expression", controller.PostExpression)
	app.Post("/recording", controller.UploadRecording)

	app.Put("/invite", controller.AcceptInvite)
	app.Put("/notification/:id", controller.UpdateSeenNotification)

	app.Delete("/account", controller.DeleteAccount)
	app.Delete("/device", controller.DeleteDevice)
	app.Delete("/friend/:id", controller.RemoveFriend)
	app.Delete("/invite/:id", controller.RemoveInvite)
	app.Delete("/emotion/:id", controller.RemoveEmotion)
	app.Delete("/expression", controller.RemoveExpression)

	return app
}
