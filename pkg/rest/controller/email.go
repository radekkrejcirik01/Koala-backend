package controller

import (
	"github.com/gofiber/fiber/v2"
)

// SendPasswordResetEmail POST /password-reset
// Disabled: it reset any account to a shared password and emailed it to an address chosen by the caller.
// Kept so old app versions get an answer, the new reset flow replaces it.
// Not 200, old app versions treat any 200 body as a sent email.
func SendPasswordResetEmail(c *fiber.Ctx) error {
	return c.Status(fiber.StatusServiceUnavailable).JSON(Response{
		Status:  "error",
		Message: "Password reset is temporarily unavailable, please contact support",
	})
}
