package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
	"github.com/radekkrejcirik01/Koala-backend/pkg/middleware"
	profilephoto "github.com/radekkrejcirik01/Koala-backend/pkg/model/profile-photo"
)

// UploadProfilePhoto POST /profile-photo
func UploadProfilePhoto(c *fiber.Ctx) error {
	username, err := middleware.Authorize(c)
	if err != nil {
		return err
	}

	t := &profilephoto.ProfilePhoto{}

	if err := c.BodyParser(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	photoUrl, err := profilephoto.UploadProfilePhoto(database.DB, username, t)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(UploadProfilePhotoResponse{
		Status:  "success",
		Message: "Profile photo successfully uploaded",
		Data:    photoUrl,
	})
}
