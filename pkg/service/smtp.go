package service

import (
	"errors"
	"fmt"
	"net/smtp"

	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
	"github.com/radekkrejcirik01/Koala-backend/pkg/middleware"
	"gorm.io/gorm"
)

type ResetPasswordEmail struct {
	Email          string
	Username       string
	FriendUsername string
}

type Support struct {
	Username string
	Message  string
	Email    string
}

type Feedback struct {
	Username string
	Message  string
}

func SendPasswordResetEmail(db *gorm.DB, t *ResetPasswordEmail) error {
	var accepted int

	if err := db.
		Table("invites").
		Select("accepted").
		Where("(sender = ? AND receiver = ?) OR (sender = ? AND receiver = ?)",
			t.Username, t.FriendUsername, t.FriendUsername, t.Username).
		Find(&accepted).
		Error; err != nil {
		return err
	}

	if accepted == 1 {
		password := database.GetResetPassword()
		newPassword := middleware.GetHashPassword(password)

		if err := db.
			Table("users").
			Where("username = ?", t.Username).
			Update("password", newPassword).
			Error; err != nil {
			return err
		}

		baseEmail, basePassword := database.GetEmailCredentials()
		auth := smtp.PlainAuth("", baseEmail, basePassword, "smtp.gmail.com")

		body := fmt.Sprintf("Hi,\nhere is your new password for your Koala Messenger account:\n\nusername: " + t.Username + "\npassword: password1234\n\nPlease change the password in Profile -> Account -> Change password\n\nHave a nice day,\nKoala Team")

		msg := []byte("To: " + t.Email + "\r\n" +
			"Subject: Koala New Password" + "\r\n" +
			"\r\n" +
			body + "\r\n")

		return smtp.SendMail("smtp.gmail.com:587", auth, baseEmail, []string{t.Email}, msg)
	}
	return errors.New("incorrect")
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
