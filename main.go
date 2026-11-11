package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	fiberadapter "github.com/awslabs/aws-lambda-go-api-proxy/fiber"
	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/chats"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/devices"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/emotions"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/invites"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/notifications"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/users"
	"github.com/radekkrejcirik01/Koala-backend/pkg/rest"
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
		&chats.Chat{},
		&chats.Message{},
	); err != nil {
		log.Fatal(err)
	}
}

// Handler will deal with Fiber working with Lambda
func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return fiberLambda.ProxyWithContext(ctx, request)
}

func main() {
	// Make the handler available for Remote Procedure Call by AWS Lambda
	lambda.Start(Handler)
}
