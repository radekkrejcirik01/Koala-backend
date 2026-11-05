package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
	"github.com/radekkrejcirik01/Koala-backend/pkg/middleware"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/emotions"
)

// GetEmotions GET /emotions?from=&to=
func GetEmotions(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}
	from := c.Query("from")
	to := c.Query("to")

	e, err := emotions.GetEmotions(database.DB, username, from, to)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(EmotionsResponse{
		Status:  "success",
		Message: "Emotions successfully get",
		Data:    e,
	})
}
