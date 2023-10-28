package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
	"github.com/radekkrejcirik01/Koala-backend/pkg/middleware"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/users"
)

// CreateUser POST /user
func CreateUser(c *fiber.Ctx) error {
	t := &users.User{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	err := users.CreateUser(database.DB, t)
	if err != nil {
		status := fiber.StatusInternalServerError

		if err.Error() == "user already exists" {
			status = fiber.StatusOK
		}

		return c.Status(status).JSON(Response{
			Status:  "error",
			Message: err.Error(),
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
		Message: "User successfully created",
		Token:   token,
	})
}

// LoginUser POST /login
func LoginUser(c *fiber.Ctx) error {
	t := &users.Login{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if err := users.LoginUser(database.DB, t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	token, err := middleware.CreateJWT(t.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(AuthResponse{
		Status:  "success",
		Message: "User successfully authenticated",
		Token:   token,
	})
}

// GetUser GET /user
func GetUser(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	user, emotions, getErr := users.GetUser(database.DB, username)
	if getErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: getErr.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(UserResponse{
		Status:   "success",
		Message:  "User successfully got",
		Data:     user,
		Emotions: emotions,
	})
}
