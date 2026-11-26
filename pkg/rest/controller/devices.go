package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
	"github.com/radekkrejcirik01/Koala-backend/pkg/middleware"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/devices"
)

// SaveDevice POST /device
func SaveDevice(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	t := &devices.Device{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if err := devices.SaveDevice(database.DB, username, t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Device successfully saved",
	})
}

// DeleteDevice DELETE /device
func DeleteDevice(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	// Old app versions send no body
	t := &devices.DeleteRequest{}
	if len(c.Body()) > 0 {
		if err := c.BodyParser(t); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(Response{
				Status:  "error",
				Message: err.Error(),
			})
		}
	}

	if err := devices.DeleteDevice(database.DB, username, t.DeviceToken); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Device successfully deleted",
	})
}
