package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
	"github.com/radekkrejcirik01/Koala-backend/pkg/middleware"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/invites"
)

// SendInvite POST /invite
func SendInvite(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	t := &invites.Invite{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	t.Sender = username

	message, err := invites.SendInvite(database.DB, t)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: message,
	})
}

// AcceptInvite PUT /invite
func AcceptInvite(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	t := &invites.Invite{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	t.Sender = username

	message, err := invites.AcceptInvite(database.DB, t)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: message,
	})
}

// GetFriends GET /friends
func GetFriends(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	friends, err := invites.GetFriends(database.DB, username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(FriendsResponse{
		Status:  "success",
		Message: "Friends successfully got",
		Data:    friends,
	})
}

// GetFriendRequests GET /friend-requests
func GetFriendRequests(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	friendRequests, err := invites.GetFriendRequests(database.DB, username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(FriendsResponse{
		Status:  "success",
		Message: "Friend requests successfully got",
		Data:    *friendRequests,
	})
}

// GetInvites GET /invites
func GetInvites(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	invites, err := invites.GetInvites(database.DB, username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(InvitesResponse{
		Status:  "success",
		Message: "Friend invites successfully got",
		Data:    invites,
	})
}

// RemoveFriend DELETE /friend/:id
func RemoveFriend(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	id := c.Params("id")

	if err := invites.RemoveFriend(database.DB, id, username); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Friend removed",
	})
}

// RemoveInvite DELETE /invite/:id
func RemoveInvite(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	id := c.Params("id")

	if err := invites.RemoveInvite(database.DB, id, username); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Inivte removed",
	})
}
