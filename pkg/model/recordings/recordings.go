package recordings

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
)

type Recording struct {
	Buffer   string
	Platform string
}

func UploadRecording(t *Recording, username string) (string, error) {
	accessKey, secretAccessKey := database.GetCredentials()
	fileName := getRecordingFileName(t.Platform)
	contentType := getContentType(t.Platform)

	sess := session.Must(session.NewSession(
		&aws.Config{
			Region: aws.String("eu-central-1"),
			Credentials: credentials.NewStaticCredentials(
				accessKey,
				secretAccessKey,
				"", // a token will be created when the session it's used.
			),
		}))

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)

	decode, _ := base64.StdEncoding.DecodeString(t.Buffer)
	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String("koala-bucket-records"),
		Key:         aws.String("recordings/" + username + "/" + fileName),
		Body:        bytes.NewReader(decode),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", err
	}

	return result.Location, nil
}

func getRecordingFileName(platform string) string {
	// Get current timestamp
	timestamp := time.Now().UnixNano()

	// Convert timestamp to string
	timestampString := fmt.Sprintf("%d", timestamp)

	// Get file extension based on platform
	var extenstion string
	if platform == "ios" {
		extenstion = ".m4a"
	} else {
		extenstion = ".mp3"
	}

	return timestampString + extenstion
}

func getContentType(platform string) string {
	if platform == "ios" {
		return "audio/mp4"
	}
	return "audio/mpeg"
}
