package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/radekkrejcirik01/Koala-backend/pkg/rest/controller"
)

// Create new REST API server
func Create() *fiber.App {
	app := fiber.New()

	app.Get("/", controller.GetIndex)
	app.Get("/user", controller.GetUser)

	app.Post("/user", controller.CreateUser)
	app.Post("/login", controller.LoginUser)

	return app
}
