package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
	"github.com/radekkrejcirik01/Koala-backend/pkg/service"
)

// SendPasswordResetEmail POST /password-reset
func SendPasswordResetEmail(c *fiber.Ctx) error {
	t := &service.ResetPasswordEmail{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if err := service.SendPasswordResetEmail(database.DB, t); err != nil {
		status := fiber.StatusInternalServerError

		if err.Error() == "incorrect" {
			status = fiber.StatusOK
		}

		return c.Status(status).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Email successfully sent",
	})
}
