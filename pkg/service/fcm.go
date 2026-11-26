package service

import (
	"context"
	"log"
	"sync"

	"firebase.google.com/go/v4/messaging"
	fcm "github.com/appleboy/go-fcm"
	"gorm.io/gorm"
)

// FCM accepts at most 500 messages in one batch
const maxBatchSize = 500

// fcmClient is created once and reused while the Lambda instance stays warm
var (
	fcmClient   *fcm.Client
	fcmClientMu sync.Mutex
)

type FcmNotification struct {
	Data    map[string]string
	Title   string
	Body    string
	Sound   string
	Devices []string
}

func GetTokensByUserId(db *gorm.DB, id int64) ([]string, error) {
	var tokens []string
	err := db.
		Table("devices").
		Select("device_token").
		Distinct("device_token").
		Where("user_id = ?", id).
		Find(&tokens).
		Error

	return tokens, err
}

func GetTokensByUserIds(db *gorm.DB, ids []int64) ([]string, error) {
	var tokens []string
	err := db.
		Table("devices").
		Select("device_token").
		Distinct("device_token").
		Where("user_id IN ?", ids).
		Find(&tokens).
		Error

	return tokens, err
}

// SendNotification sends the notification to all devices in parallel batches,
// failures are only logged so they never fail the request,
// tokens of uninstalled apps are deleted
func SendNotification(db *gorm.DB, t *FcmNotification) error {
	if len(t.Devices) == 0 {
		return nil
	}

	ctx := context.Background()

	client, err := getClient(ctx)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	apnsConfig := NewAPNSConfig(t.Sound, 1)
	androidConfig := NewAndroidConfig(t.Sound)

	var messages []*messaging.Message
	for _, token := range t.Devices {
		messages = append(messages, &messaging.Message{
			Token: token,
			Data:  t.Data,
			Notification: &messaging.Notification{
				Title: t.Title,
				Body:  t.Body,
			},
			APNS:    apnsConfig,
			Android: androidConfig,
		})
	}

	var unregistered []string
	for start := 0; start < len(messages); start += maxBatchSize {
		end := min(start+maxBatchSize, len(messages))

		response, err := client.Send(ctx, messages[start:end]...)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		log.Printf("FCM sent: %d, failed: %d\n", response.SuccessCount, response.FailureCount)
		// Responses are in the order of the batch's messages
		for i, r := range response.Responses {
			if r.Error == nil {
				continue
			}
			log.Println(r.Error.Error())
			if messaging.IsUnregistered(r.Error) {
				unregistered = append(unregistered, messages[start+i].Token)
			}
		}
	}

	if len(unregistered) > 0 {
		if err := db.
			Exec("DELETE FROM devices WHERE device_token IN ?", unregistered).
			Error; err != nil {
			log.Println(err.Error())
		}
	}

	return nil
}

func getClient(ctx context.Context) (*fcm.Client, error) {
	fcmClientMu.Lock()
	defer fcmClientMu.Unlock()

	if fcmClient != nil {
		return fcmClient, nil
	}

	client, err := fcm.NewClient(ctx, fcm.WithCredentialsFile("./fcm-client.json"))
	if err != nil {
		return nil, err
	}

	fcmClient = client

	return fcmClient, nil
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
