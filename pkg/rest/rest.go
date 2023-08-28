package rest

import "github.com/gofiber/fiber/v2"

// Create new REST API server
func Create() *fiber.App {
	app := fiber.New()

	return app
}
