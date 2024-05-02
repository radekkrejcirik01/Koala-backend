package service

import (
	"fmt"
	"net/smtp"

	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
)

type ResetPasswordEmail struct {
	Email    string
	Username string
	Friends  string
}

// SendPasswordResetEmail send e
func SendPasswordResetEmail(t *ResetPasswordEmail) error {
	baseEmail, basePassword := database.GetEmailCredentials()

	auth := smtp.PlainAuth("", baseEmail, basePassword, "smtp.gmail.com")

	body := fmt.Sprintf("Username: %s\nEmail: %s\nFriends: %s\n", t.Username, t.Email, t.Friends)

	msg := []byte("To: " + t.Email + "\r\n" +
		"Subject: Koala Account Password" + "\r\n" +
		"\r\n" +
		body + "\r\n")

	return smtp.SendMail("smtp.gmail.com:587", auth, baseEmail, []string{baseEmail}, msg)
}
