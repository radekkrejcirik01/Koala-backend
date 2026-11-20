package service

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
)

// sendTimeout limits how long a request waits for the sockets, events are best effort
const sendTimeout = 5 * time.Second

// wsClient is created once and reused while the Lambda instance stays warm,
// it posts to the WebSocket API at WS_ENDPOINT (https://{id}.execute-api.{region}.amazonaws.com/{stage})
var (
	wsClient     *apigatewaymanagementapi.ApiGatewayManagementApi
	wsClientOnce sync.Once
)

type SocketMessage struct {
	ConnectionId string
	Data         []byte
}

// SendToConnections posts the messages in parallel and returns the connections which are gone,
// other failures are only logged so they never fail the request
func SendToConnections(messages []SocketMessage) []string {
	client := getWsClient()
	if client == nil || len(messages) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), sendTimeout)
	defer cancel()

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		gone []string
	)
	for _, message := range messages {
		wg.Add(1)
		go func(message SocketMessage) {
			defer wg.Done()

			_, err := client.PostToConnectionWithContext(ctx, &apigatewaymanagementapi.PostToConnectionInput{
				ConnectionId: aws.String(message.ConnectionId),
				Data:         message.Data,
			})
			if err == nil {
				return
			}

			var awsErr awserr.Error
			if errors.As(err, &awsErr) && awsErr.Code() == apigatewaymanagementapi.ErrCodeGoneException {
				mu.Lock()
				gone = append(gone, message.ConnectionId)
				mu.Unlock()
				return
			}

			log.Println(err.Error())
		}(message)
	}
	wg.Wait()

	return gone
}

func getWsClient() *apigatewaymanagementapi.ApiGatewayManagementApi {
	wsClientOnce.Do(func() {
		endpoint := os.Getenv("WS_ENDPOINT")
		if endpoint == "" {
			log.Println("WS_ENDPOINT is not set, socket events are not sent")
			return
		}

		sess := session.Must(session.NewSession(
			&aws.Config{
				Region: aws.String("eu-central-1"),
			}))

		wsClient = apigatewaymanagementapi.New(sess, aws.NewConfig().WithEndpoint(endpoint))
	})

	return wsClient
}
