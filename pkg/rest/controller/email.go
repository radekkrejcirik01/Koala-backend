package controller

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
	"github.com/radekkrejcirik01/Koala-backend/pkg/middleware"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/codes"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/users"
)

// SendPasswordResetEmail POST /password-reset
// Disabled: it reset any account to a shared password and emailed it to an address chosen by the caller.
// Kept so old app versions get an answer, /password-reset/request and /password-reset/confirm replace it.
// Not 200, old app versions treat any 200 body as a sent email.
func SendPasswordResetEmail(c *fiber.Ctx) error {
	return c.Status(fiber.StatusServiceUnavailable).JSON(Response{
		Status:  "error",
		Message: "Password reset is temporarily unavailable, please contact support",
	})
}

// RequestPasswordReset POST /password-reset/request
// Same answer whether the account exists and has an email or not
func RequestPasswordReset(c *fiber.Ctx) error {
	t := &users.Username{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if err := users.RequestPasswordReset(database.DB, t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "If the account has a verified email, a code was sent to it",
	})
}

// ResetPassword POST /password-reset/confirm
// Logs out all devices and returns a new token
func ResetPassword(c *fiber.Ctx) error {
	t := &users.PasswordReset{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	tokenVersion, err := users.ResetPassword(database.DB, t)
	if err != nil {
		status := fiber.StatusInternalServerError

		if errors.Is(err, codes.ErrInvalid) || t.NewPassword == "" {
			status = fiber.StatusBadRequest
		}

		return c.Status(status).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	token, err := middleware.CreateJWT(t.Username, tokenVersion)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(AuthResponse{
		Status:  "success",
		Message: "Password successfully reset",
		Token:   token,
	})
}
