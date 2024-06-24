package service

import (
	"context"
	"log"

	"firebase.google.com/go/v4/messaging"
	fcm "github.com/appleboy/go-fcm"
	"gorm.io/gorm"
)

type FcmNotification struct {
	Data    map[string]string
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
		Distinct().
		Where("username = ?", username).
		Find(&tokens).
		Error

	return tokens, err
}

func GetTokensByUserId(db *gorm.DB, id int64) ([]string, error) {
	var tokens []string
	err := db.
		Table("devices").
		Select("device_token").
		Distinct().
		Where("user_id = ?", id).
		Find(&tokens).
		Error

	return tokens, err
}

func GetTokensByUsernames(db *gorm.DB, usernames []string) ([]string, error) {
	var tokens []string
	err := db.
		Table("devices").
		Select("device_token").
		Distinct().
		Where("username IN ?", usernames).
		Find(&tokens).Error

	return tokens, err
}

func GetTokensByUserIds(db *gorm.DB, ids []int64) ([]string, error) {
	var tokens []string
	err := db.
		Table("devices").
		Select("device_token").
		Distinct().
		Where("user_id IN ?", ids).
		Find(&tokens).
		Error

	return tokens, err
}

func SendNotification(t *FcmNotification) error {
	ctx := context.Background()

	client, err := fcm.NewClient(ctx, fcm.WithCredentialsFile("./fcm-client.json"))
	if err != nil {
		log.Println(err.Error())
	}

	apnsConfig := NewAPNSConfig(t.Sound, 1)
	androidConfig := NewAndroidConfig(t.Sound)

	for _, token := range t.Devices {
		msg := messaging.Message{
			Token: token,
			Data:  t.Data,
			Notification: &messaging.Notification{
				Title: t.Title,
				Body:  t.Body,
			},
			APNS:    apnsConfig,
			Android: androidConfig,
		}

		response, err := client.Send(ctx, &msg)
		if err != nil {
			log.Println(err.Error())
		}
		log.Printf("%#v\n", response)
	}

	return nil
}

func NewAPNSConfig(sound string, badge int) *messaging.APNSConfig {
	return &messaging.APNSConfig{
		Payload: &messaging.APNSPayload{
			Aps: &messaging.Aps{
				Sound: sound,
				Badge: &badge,
			},
		},
	}
}

func NewAndroidConfig(sound string) *messaging.AndroidConfig {
	return &messaging.AndroidConfig{
		Notification: &messaging.AndroidNotification{
			Priority: messaging.PriorityHigh,
			Sound:    sound,
		},
	}
}
