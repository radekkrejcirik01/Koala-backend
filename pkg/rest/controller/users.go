package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
	"github.com/radekkrejcirik01/Koala-backend/pkg/middleware"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/users"
)

// GetUser GET /user
func GetIndex(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Index successfully got",
	})
}

// CreateUser POST /user
func CreateUser(c *fiber.Ctx) error {
	t := &users.User{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	message, createErr := users.CreateUser(database.DB, t)

	if createErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: createErr.Error(),
		})
	}

	token, tokenErr := middleware.CreateJWT(t.Username)
	if tokenErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": tokenErr.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(AuthResponse{
		Status:  "success",
		Message: message,
		Token:   token,
	})
}

// LoginUser POST /login
func LoginUser(c *fiber.Ctx) error {
	t := &users.Login{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if err := users.LoginUser(database.DB, t); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(AuthResponse{
		Status:  "success",
		Message: "User successfully authenticated",
		Token:   "",
	})
}

// GetUser GET /user
func GetUser(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	user, getErr := users.GetUser(database.DB, username)
	if getErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: getErr.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(UserResponse{
		Status:  "success",
		Message: "User successfully got",
		Data:    user,
	})
}
