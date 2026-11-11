package service

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
)

// s3Uploader is created once and reused while the Lambda instance stays warm
var (
	s3Uploader     *s3manager.Uploader
	s3UploaderOnce sync.Once
)

func GetS3Uploader() *s3manager.Uploader {
	s3UploaderOnce.Do(func() {
		accessKey, secretAccessKey := database.GetCredentials()

		sess := session.Must(session.NewSession(
			&aws.Config{
				Region: aws.String("eu-central-1"),
				Credentials: credentials.NewStaticCredentials(
					accessKey,
					secretAccessKey,
					"", // a token will be created when the session it's used.
				),
			}))

		s3Uploader = s3manager.NewUploader(sess)
	})

	return s3Uploader
}
