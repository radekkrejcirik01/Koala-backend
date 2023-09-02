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

// SendSupportNotification POST /support-notification
func SendSupportNotification(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	t := &notifications.SupportNotification{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if err := notifications.SendSupportNotification(database.DB, t, username); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Support notification successfully sent",
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

// GetHistory GET /history
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
