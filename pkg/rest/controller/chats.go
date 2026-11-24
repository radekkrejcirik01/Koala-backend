package controller

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
	"github.com/radekkrejcirik01/Koala-backend/pkg/middleware"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/chats"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/settings"
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

// GetChatMessages GET /chats/:friendId/messages?lastId=&firstId=&aroundId=
func GetChatMessages(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}
	friendId := c.Params("friendId")
	lastId := c.Query("lastId")
	firstId := c.Query("firstId")
	aroundId := c.Query("aroundId")

	data, lastReadId, err := chats.GetMessages(database.DB, username, friendId, lastId, firstId, aroundId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(ChatMessagesResponse{
		Status:     "success",
		Message:    "Messages successfully got",
		Data:       data,
		LastReadId: lastReadId,
	})
}

// SearchChatMessages GET /chats/:friendId/messages/search?q=&lastId=
func SearchChatMessages(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}
	friendId := c.Params("friendId")
	q := c.Query("q")
	lastId := c.Query("lastId")

	data, err := chats.SearchMessages(database.DB, username, friendId, q, lastId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(ChatMessagesResponse{
		Status:  "success",
		Message: "Messages successfully found",
		Data:    data,
	})
}

// GetChatEmotions GET /chats/:friendId/emotions?senderId=&lastId=&from=
func GetChatEmotions(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}
	friendId := c.Params("friendId")
	senderId := c.Query("senderId")
	lastId := c.Query("lastId")
	from := c.Query("from")

	data, err := chats.GetEmotions(database.DB, username, friendId, senderId, lastId, from)
	if errors.Is(err, chats.ErrNotFriend) {
		return c.Status(fiber.StatusForbidden).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(ChatEmotionsResponse{
		Status:      "success",
		Message:     "Emotions successfully got",
		Data:        data.Messages,
		UserCount:   data.UserCount,
		FriendCount: data.FriendCount,
	})
}

// GetChatSettings GET /chats/:friendId/settings
func GetChatSettings(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}
	friendId := c.Params("friendId")

	data, err := chats.GetSettings(database.DB, username, friendId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(ChatSettingsResponse{
		Status:  "success",
		Message: "Settings successfully got",
		Data:    data,
	})
}

// UpdateChatSettings PUT /chats/:friendId/settings
func UpdateChatSettings(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}
	friendId := c.Params("friendId")

	t := &settings.UpdateRequest{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	err = chats.UpdateSettings(database.DB, username, friendId, t)
	if errors.Is(err, chats.ErrNotFriend) {
		return c.Status(fiber.StatusForbidden).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Settings successfully updated",
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
