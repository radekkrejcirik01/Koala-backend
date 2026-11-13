package service

import (
	"fmt"
	"net/mail"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sesv2"
)

const fromEmail = "Koala <no-reply@koalasharing.com>"
const supportEmail = "support@koalasharing.com"

type Support struct {
	Username string
	Message  string
	Email    string
}

type Feedback struct {
	Username string
	Message  string
}

// sesClient is created once and reused while the Lambda instance stays warm,
// it uses the Lambda role, which may send only from @koalasharing.com
var (
	sesClient     *sesv2.SESV2
	sesClientOnce sync.Once
)

func getSESClient() *sesv2.SESV2 {
	sesClientOnce.Do(func() {
		sess := session.Must(session.NewSession(
			&aws.Config{
				Region: aws.String("eu-central-1"),
			}))

		sesClient = sesv2.New(sess)
	})

	return sesClient
}

func sendEmail(to, subject, body string, replyTo []string) error {
	_, err := getSESClient().SendEmail(&sesv2.SendEmailInput{
		FromEmailAddress: aws.String(fromEmail),
		Destination: &sesv2.Destination{
			ToAddresses: aws.StringSlice([]string{to}),
		},
		ReplyToAddresses: aws.StringSlice(replyTo),
		Content: &sesv2.EmailContent{
			Simple: &sesv2.Message{
				Subject: &sesv2.Content{Data: aws.String(subject)},
				Body: &sesv2.Body{
					Text: &sesv2.Content{Data: aws.String(body)},
				},
			},
		},
	})

	return err
}

// SendSupport sends the message to support, replying to it goes to the user's email
func SendSupport(t *Support) error {
	var replyTo []string
	if address, err := mail.ParseAddress(t.Email); err == nil {
		replyTo = []string{address.Address}
	}

	body := fmt.Sprintf("Username: %s\nEmail: %s\nMessage: %s\n", t.Username, t.Email, t.Message)

	return sendEmail(supportEmail, "Koala Support", body, replyTo)
}

func SendFeedback(t *Feedback) error {
	body := fmt.Sprintf("Username: %s\nFeedback: %s\n", t.Username, t.Message)

	return sendEmail(supportEmail, "Koala Feedback", body, nil)
}
