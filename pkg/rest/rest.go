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
	app.Get("/friend-requests", controller.GetFriendRequests)
	app.Get("/notifications/:lastId?", controller.GetNotifications)
	app.Get("/unseen-notifications", controller.GetUnseenNotifications)
	app.Get("/track/:lastId?", controller.GetTrack)

	app.Post("/user", controller.CreateUser)
	app.Post("/login", controller.LoginUser)
	app.Post("/invite", controller.SendInvite)
	app.Post("/device", controller.SaveDevice)
	app.Post("/emotion-notification", controller.SendEmotionNotification)
	app.Post("/support-notification", controller.SendSupportNotification)

	app.Put("/invite", controller.AcceptInvite)

	app.Delete("/account", controller.DeleteAccount)
	app.Delete("/device", controller.DeleteDevice)

	return app
}
