package recordings

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/radekkrejcirik01/Koala-backend/pkg/service"
)

func UploadRecording(buffer string, userId int64) (string, error) {
	id := strconv.Itoa(int(userId))

	fileName := getFileName()

	uploader := service.GetS3Uploader()

	decode, _ := base64.StdEncoding.DecodeString(buffer)
	key := "recordings/" + id + "/" + fileName
	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String("koala-bucket-voice-messages"),
		Key:         aws.String(key),
		Body:        bytes.NewReader(decode),
		ContentType: aws.String("audio/mpeg"),
	})
	if err != nil {
		return "", err
	}

	return service.GetMediaUrl(key, result.Location), nil
}

func getFileName() string {
	const extension = ".mp3"

	// Get current timestamp
	timestamp := time.Now().UnixNano()

	// Convert timestamp to string
	fileName := fmt.Sprintf("%d", timestamp)

	return fileName + extension
}
