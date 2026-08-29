package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
	"github.com/radekkrejcirik01/Koala-backend/pkg/middleware"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/notifications"
)

// GetNotifications GET /notifications/:lastId?
func GetNotifications(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}
	lastId := c.Params("lastId")

	n, err := notifications.GetNotifications(database.DB, username, lastId)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(NotificationsResponse{
		Status:  "success",
		Message: "Notifications successfully get",
		Data:    n,
	})
}

// GetConversation GET /conversation/:id
func GetConversation(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}
	id := c.Params("id")

	conversation, err := notifications.GetConversation(database.DB, username, id)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(ConversationResponse{
		Status:  "success",
		Message: "Conversation successfully get",
		Data:    conversation,
	})
}

// GetUnseenNotifications GET /unseen-notifications
func GetUnseenNotifications(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	n, err := notifications.GetUnseenNotifications(database.DB, username)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(UnseenNotificationsResponse{
		Status:              "success",
		Message:             "Unseen notifications successfully get",
		UnseenNotifications: *n,
	})
}

// UpdateSeenNotification PUT /notification/:id
func UpdateSeenNotification(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	id := c.Params("id")

	if err := notifications.UpdateSeenNotification(database.DB, username, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Seen notification successfully updated",
	})
}

// UpdateNotificationReaction PUT /notification/:id/reaction
func UpdateNotificationReaction(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	id := c.Params("id")
	t := &notifications.UpdateReactionRequest{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if t.Reaction == nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "reaction is required",
		})
	}

	if err := notifications.UpdateNotificationReaction(database.DB, username, id, t.Reaction); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Notification reaction successfully updated",
	})
}
