package service

import (
	"log"

	"github.com/appleboy/go-fcm"
	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
	"gorm.io/gorm"
)

type FcmNotification struct {
	Data    map[string]interface{}
	Title   string
	Body    string
	Sound   string
	Devices []string
}

func GetTokensByUsername(db *gorm.DB, username string) ([]string, error) {
	var tokens []string
	err := db.
		Table("devices").
		Select("device_token").
		Where("username = ?", username).
		Find(&tokens).
		Error

	return tokens, err
}

func GetTokensByUsernames(db *gorm.DB, usernames []string) ([]string, error) {
	var tokens []string
	err := db.
		Table("devices").
		Select("device_token").
		Where("username IN ?", usernames).
		Find(&tokens).Error

	return tokens, err
}

func SendNotification(t *FcmNotification) error {
	fcmClient := database.GetFcmClient()
	tokens := t.Devices

	for _, token := range tokens {
		msg := &fcm.Message{
			To:   token,
			Data: t.Data,
			Notification: &fcm.Notification{
				Title: t.Title,
				Body:  t.Body,
				Badge: "1",
				Sound: t.Sound,
			},
		}

		client, err := fcm.NewClient(fcmClient)
		if err != nil {
			log.Fatalln(err)
			return err
		}

		response, err := client.Send(msg)
		if err != nil {
			log.Fatalln(err)
			return err
		}

		log.Printf("%#v\n", response)
	}

	return nil
}
