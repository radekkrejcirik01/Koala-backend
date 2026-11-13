package service

import (
	"fmt"
	"net/smtp"

	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
)

type Support struct {
	Username string
	Message  string
	Email    string
}

type Feedback struct {
	Username string
	Message  string
}

func SendSupport(t *Support) error {
	baseEmail, basePassword := database.GetEmailCredentials()

	auth := smtp.PlainAuth("", baseEmail, basePassword, "smtp.gmail.com")

	body := fmt.Sprintf("Username: %s\nEmail: %s\nMessage: %s\n", t.Username, t.Email, t.Message)

	msg := []byte("To: " + baseEmail + "\r\n" +
		"Subject: Koala Support" + "\r\n" +
		"\r\n" +
		body + "\r\n")

	return smtp.SendMail("smtp.gmail.com:587", auth, baseEmail, []string{baseEmail}, msg)
}

func SendFeedback(t *Feedback) error {
	baseEmail, basePassword := database.GetEmailCredentials()

	auth := smtp.PlainAuth("", baseEmail, basePassword, "smtp.gmail.com")

	body := fmt.Sprintf("Username: %s\nFeedback: %s\n", t.Username, t.Message)

	msg := []byte("To: " + baseEmail + "\r\n" +
		"Subject: Koala Feedback" + "\r\n" +
		"\r\n" +
		body + "\r\n")

	return smtp.SendMail("smtp.gmail.com:587", auth, baseEmail, []string{baseEmail}, msg)
}
