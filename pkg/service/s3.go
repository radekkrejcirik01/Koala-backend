package service

import (
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// s3Uploader is created once and reused while the Lambda instance stays warm,
// it uses the Lambda role, which may only upload to the media folders
var (
	s3Uploader     *s3manager.Uploader
	s3UploaderOnce sync.Once
)

func GetS3Uploader() *s3manager.Uploader {
	s3UploaderOnce.Do(func() {
		sess := session.Must(session.NewSession(
			&aws.Config{
				Region: aws.String("eu-central-1"),
			}))

		s3Uploader = s3manager.NewUploader(sess)
	})

	return s3Uploader
}

// GetMediaUrl returns the public URL of an uploaded file, served from MEDIA_URL (CloudFront) when it's set
func GetMediaUrl(key string, s3Location string) string {
	mediaUrl := os.Getenv("MEDIA_URL")
	if mediaUrl == "" {
		return s3Location
	}

	return mediaUrl + "/" + key
}
