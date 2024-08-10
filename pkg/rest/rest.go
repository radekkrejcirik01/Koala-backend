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
	app.Get("/friend-notifications/:id/:lastId?", controller.GetFriendNotifications)
	app.Get("/conversation/:id", controller.GetConversation)
	app.Get("/unseen-notifications", controller.GetUnseenNotifications)
	app.Get("/history/:lastId?", controller.GetHistory)
	app.Get("/user-history/:receiverId/:lastId?", controller.GetUserHistory)
	app.Get("/emotions-messages/:type", controller.GetEmotionsMessages)
	app.Get("/emotions", controller.GetEmotions)
	app.Get("/expressions", controller.GetExpressions)
	app.Get("/last-online/:id", controller.GetLastOnline)
	app.Get("/replies/:id", controller.GetReplies)
	app.Get("/check-on-messages/:id", controller.GetCheckOnMessages)
	app.Get("/last-shared-message", controller.GetLastSharedMessage)

	app.Post("/user", controller.CreateUser)
	app.Post("/login", controller.LoginUser)
	app.Post("/username", controller.CheckUsername)
	app.Post("/invite", controller.SendInvite)
	app.Post("/device", controller.SaveDevice)
	app.Post("/emotion-notification", controller.SendEmotionNotification)
	app.Post("/emotion-message/:type?", controller.SendEmotionMessage)
	app.Post("/status-reply-notification", controller.SendStatusReplyNotification)
	app.Post("/status-reply-message", controller.SendStatusReplyMessage)
	app.Post("/message-notification", controller.SendMessageNotification)
	app.Post("/message", controller.SendMessage)
	app.Post("/check-on-message", controller.SendCheckOnMessage)
	app.Post("/emotion", controller.AddEmotion)
	app.Post("/removed-emotion", controller.AddRemovedEmotion)
	app.Post("/expression", controller.PostExpression)
	app.Post("/password-reset", controller.SendPasswordResetEmail)
	app.Post("/reply", controller.AddReply)
	app.Post("/user-check-on-message", controller.AddCheckOnMessage)

	app.Put("/invite", controller.AcceptInvite)
	app.Put("/notification/:id", controller.UpdateSeenNotification)
	app.Put("/user-password", controller.ChangePassword)
	app.Put("/last-online", controller.UpdateLastOnline)

	app.Delete("/account", controller.DeleteAccount)
	app.Delete("/device", controller.DeleteDevice)
	app.Delete("/friend/:id", controller.RemoveFriend)
	app.Delete("/invite/:id", controller.RemoveInvite)
	app.Delete("/emotion/:id", controller.RemoveEmotion)
	app.Delete("/expression", controller.RemoveExpression)
	app.Delete("/message/:id", controller.DeleteMessage)
	app.Delete("/reply/:id", controller.DeleteReply)
	app.Delete("/check-on-message/:id", controller.DeleteCheckOnMessage)

	return app
}
