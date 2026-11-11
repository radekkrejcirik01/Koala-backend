package profilephoto

import (
	"bytes"
	"encoding/base64"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/radekkrejcirik01/Koala-backend/pkg/service"
	"gorm.io/gorm"
)

type ProfilePhoto struct {
	Buffer   string
	FileName string
}

func UploadProfilePhoto(db *gorm.DB, username string, t *ProfilePhoto) (string, error) {
	var userId int64

	if err := db.
		Table("users").
		Select("id").
		Where("username = ?", username).
		Find(&userId).
		Error; err != nil {
		return "", err
	}

	uploader := service.GetS3Uploader()

	decode, _ := base64.StdEncoding.DecodeString(t.Buffer)
	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String("koala-bucket-profile-photos"),
		Key:         aws.String("profile-photos/" + strconv.Itoa(int(userId)) + "/" + t.FileName),
		Body:        bytes.NewReader(decode),
		ContentType: aws.String("image/jpeg"),
	})
	if err != nil {
		return "", err
	}

	if err := db.
		Table("users").
		Where("id = ?", userId).
		Update("profile_photo", result.Location).
		Error; err != nil {
		return "", err
	}

	return result.Location, nil
}
