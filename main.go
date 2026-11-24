package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	fiberadapter "github.com/awslabs/aws-lambda-go-api-proxy/fiber"
	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/chats"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/codes"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/connections"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/devices"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/emotions"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/invites"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/notifications"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/settings"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/users"
	"github.com/radekkrejcirik01/Koala-backend/pkg/rest"
	"github.com/radekkrejcirik01/Koala-backend/pkg/websocket"
)

var fiberLambda *fiberadapter.FiberLambda

func init() {
	database.Connect()

	// Migrations slow down every cold start, they run only when MIGRATE=true is set
	if os.Getenv("MIGRATE") == "true" {
		migrate()
	}

	fiberLambda = fiberadapter.New(rest.Create())
}

func migrate() {
	if err := database.DB.AutoMigrate(
		&users.User{},
		&invites.Invite{},
		&devices.Device{},
		&notifications.Notification{},
		&emotions.Emotion{},
		&emotions.Recipient{},
		&chats.Chat{},
		&chats.Message{},
		&codes.Code{},
		&connections.Connection{},
		&settings.Setting{},
	); err != nil {
		log.Fatal(err)
	}
}

// eventContext tells the WebSocket API events, which have a connection id, from the REST API ones
type eventContext struct {
	RequestContext struct {
		ConnectionID string `json:"connectionId"`
	} `json:"requestContext"`
}

// Handler serves the REST API with Fiber and the WebSocket API routes from the same function
func Handler(ctx context.Context, event json.RawMessage) (events.APIGatewayProxyResponse, error) {
	var eventCtx eventContext
	if err := json.Unmarshal(event, &eventCtx); err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	if eventCtx.RequestContext.ConnectionID != "" {
		var request events.APIGatewayWebsocketProxyRequest
		if err := json.Unmarshal(event, &request); err != nil {
			return events.APIGatewayProxyResponse{}, err
		}
		return websocket.Handle(request), nil
	}

	var request events.APIGatewayProxyRequest
	if err := json.Unmarshal(event, &request); err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	return fiberLambda.ProxyWithContext(ctx, request)
}

func main() {
	// Make the handler available for Remote Procedure Call by AWS Lambda
	lambda.Start(Handler)
}
