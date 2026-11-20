package controller

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
	"github.com/radekkrejcirik01/Koala-backend/pkg/middleware"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/codes"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/connections"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/invites"
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
	token, tokenErr := middleware.CreateJWT(t.Username, 0)
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

	tokenVersion, err := users.LoginUser(database.DB, t)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	token, err := middleware.CreateJWT(t.Username, tokenVersion)
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

	user, err := users.GetUser(database.DB, username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(UserResponse{
		Status:  "success",
		Message: "User successfully got",
		Data:    user,
	})
}

// CheckUsername POST /username
func CheckUsername(c *fiber.Ctx) error {
	t := &users.Username{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if err := users.CheckUsername(database.DB, t.Username); err != nil {
		status := fiber.StatusInternalServerError

		if err.Error() == "username is already taken" {
			status = fiber.StatusOK
		}

		return c.Status(status).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Username successfully checked",
	})
}

// ChangeName PUT /user-name
func ChangeName(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	t := &users.Name{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if err := users.ChangeName(database.DB, username, t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Name successfully changed",
	})
}

// ChangePassword PUT /user-password
func ChangePassword(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	t := &users.Password{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if err := users.ChangePassword(database.DB, username, t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Password successfully changed",
	})
}

// ChangeEmail PUT /user-email
func ChangeEmail(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	t := &users.Email{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if err := users.ChangeEmail(database.DB, username, t); err != nil {
		status := fiber.StatusInternalServerError

		if errors.Is(err, users.ErrInvalidEmail) || errors.Is(err, codes.ErrTooSoon) {
			status = fiber.StatusBadRequest
		}

		return c.Status(status).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Verification code successfully sent",
	})
}

// VerifyEmail PUT /user-email/verify
func VerifyEmail(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	t := &users.EmailCode{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if err := users.VerifyEmail(database.DB, username, t); err != nil {
		status := fiber.StatusInternalServerError

		if errors.Is(err, codes.ErrInvalid) {
			status = fiber.StatusBadRequest
		}

		return c.Status(status).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Email successfully verified",
	})
}

// UpdateLastOnline PUT /last-online
func UpdateLastOnline(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	if err := users.UpdateLastOnline(database.DB, username); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Successfully updated online time",
	})
}

// GetLastOnline GET /last-online/:id, only friends can see user's online status
func GetLastOnline(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	id := c.Params("id")

	userId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	isFriend, err := invites.IsFriend(database.DB, username, userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}
	if !isFriend {
		return c.Status(fiber.StatusForbidden).JSON(Response{
			Status:  "error",
			Message: "user is not your friend",
		})
	}

	time, err := users.GetLastOnline(database.DB, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	onlineIds, err := connections.GetOnlineUserIds(database.DB, []int64{userId})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(LastOnlineResponse{
		Status:  "success",
		Message: "Successfully got last online time",
		Time:    time,
		Online:  len(onlineIds) > 0,
	})
}
