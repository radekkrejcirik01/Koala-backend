package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
	"github.com/radekkrejcirik01/Koala-backend/pkg/middleware"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/notifications"
)

// SendEmotionNotification POST /emotion-notification
func SendEmotionNotification(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	t := &notifications.EmotionNotification{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if err := notifications.SendEmotionNotification(database.DB, t, username); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Emotion notification successfully sent",
	})
}

// SendStatusReplyNotification POST /status-reply-notification
func SendStatusReplyNotification(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	t := &notifications.StatusReplyNotification{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if err := notifications.SendStatusReplyNotification(database.DB, t, username); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Status reply notification successfully sent",
	})
}

// SendMessageNotification POST /message-notification
func SendMessageNotification(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	t := &notifications.MessageNotification{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if err := notifications.SendMessageNotification(database.DB, t, username); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Message notification successfully sent",
	})
}

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

// GetHistory GET /history/:lastId?
func GetHistory(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}
	lastId := c.Params("lastId")

	history, err := notifications.GetHistory(database.DB, username, lastId)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(HistoryResponse{
		Status:  "success",
		Message: "History successfully get",
		Data:    history,
	})
}

// GetUserHistory GET /user-history/:receiverId/:lastId?
func GetUserHistory(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	receiverId := c.Params("receiverId")
	lastId := c.Params("lastId")

	history, err := notifications.GetUserHistory(database.DB, username, receiverId, lastId)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(HistoryResponse{
		Status:  "success",
		Message: "User history successfully get",
		Data:    history,
	})
}
