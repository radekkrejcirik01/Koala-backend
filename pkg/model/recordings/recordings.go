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

func UploadRecording(buffer, username string) (string, error) {
	accessKey, secretAccessKey := database.GetCredentials()

	fileName := getFileName()

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

	decode, _ := base64.StdEncoding.DecodeString(buffer)
	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String("koala-bucket-voice-messages"),
		Key:         aws.String("recordings/" + username + "/" + fileName),
		Body:        bytes.NewReader(decode),
		ContentType: aws.String("audio/mpeg"),
	})
	if err != nil {
		return "", err
	}

	return result.Location, nil
}

func getFileName() string {
	const extension = ".mp3"

	// Get current timestamp
	timestamp := time.Now().UnixNano()

	// Convert timestamp to string
	fileName := fmt.Sprintf("%d", timestamp)

	return fileName + extension
}
