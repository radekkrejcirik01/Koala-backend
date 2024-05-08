package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
	"github.com/radekkrejcirik01/Koala-backend/pkg/middleware"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/replies"
)

func AddReply(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	t := &replies.Reply{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if err := replies.AddReply(database.DB, t, username); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Reply successfully added",
	})
}

func GetReplies(c *fiber.Ctx) error {
	_, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	id := c.Params("id")

	replies, err := replies.GetReplies(database.DB, id)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(RepliesResponse{
		Status:  "success",
		Message: "Replies successfully get",
		Data:    replies,
	})
}

func DeleteReply(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	id := c.Params("id")

	if err := replies.DeleteReply(database.DB, username, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Reply successfully deleted",
	})
}
