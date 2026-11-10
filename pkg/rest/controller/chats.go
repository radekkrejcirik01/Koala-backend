package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
	"github.com/radekkrejcirik01/Koala-backend/pkg/middleware"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/chats"
)

// GetChats GET /chats?lastTimeMs=&lastId=
func GetChats(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}
	lastTimeMs := c.Query("lastTimeMs")
	lastId := c.Query("lastId")

	data, err := chats.GetChats(database.DB, username, lastTimeMs, lastId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(ChatsResponse{
		Status:  "success",
		Message: "Chats successfully got",
		Data:    data,
	})
}

// GetUnreadChats GET /chats/unread
func GetUnreadChats(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	count, err := chats.GetUnreadChats(database.DB, username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(UnreadChatsResponse{
		Status:      "success",
		Message:     "Unread chats successfully got",
		UnreadChats: count,
	})
}

// GetChatMessages GET /chats/:friendId/messages?lastId=
func GetChatMessages(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}
	friendId := c.Params("friendId")
	lastId := c.Query("lastId")

	data, err := chats.GetMessages(database.DB, username, friendId, lastId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(ChatMessagesResponse{
		Status:  "success",
		Message: "Messages successfully got",
		Data:    data,
	})
}

// ReadChat PUT /chats/:friendId/read
func ReadChat(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}
	friendId := c.Params("friendId")

	if err := chats.ReadChat(database.DB, username, friendId); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Chat marked as read",
	})
}

// SendChatMessage POST /chats/:friendId/messages
func SendChatMessage(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}
	friendId := c.Params("friendId")

	t := &chats.SendMessageRequest{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	message, err := chats.SendMessage(database.DB, username, friendId, t)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(ChatMessageResponse{
		Status:  "success",
		Message: "Message successfully sent",
		Data:    message,
	})
}

// SendChatEmotion POST /chats/emotion
func SendChatEmotion(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	t := &chats.SendEmotionRequest{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if err := chats.SendEmotion(database.DB, username, t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Emotion successfully sent",
	})
}

// UpdateChatReaction PUT /chats/messages/:id/reaction
func UpdateChatReaction(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}
	id := c.Params("id")

	t := &chats.ReactionRequest{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if err := chats.UpdateReaction(database.DB, username, id, t.Reaction); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Reaction successfully updated",
	})
}

// DeleteChatMessage DELETE /chats/messages/:id
func DeleteChatMessage(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}
	id := c.Params("id")

	if err := chats.DeleteMessage(database.DB, username, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Message successfully deleted",
	})
}
