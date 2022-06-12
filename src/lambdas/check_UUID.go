package main

import (
	"common"
	"context"
	"dynamoDAO"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	"log"
	"time"
)

// checks DB if uuid exist if not creates a new one

// TODO - Test handler function using SAM

func Handler(ctx context.Context, sqsEvent events.SQSEvent) {

	cfg, _ := config.LoadDefaultConfig(context.TODO())
	dynamodbClient := dynamodb.NewFromConfig(cfg)

	msName := sqsEvent.Records[0].Body

	var UUID = ""
	UUID, _ = dynamoDAO.Get(dynamodbClient, common.TableName, msName)
	log.Println("UUID for ", msName, ":", UUID)
	if UUID == "" {
		UUID = uuid.New().String()
		dynamoDAO.Put(dynamodbClient, common.TableName, msName, UUID, time.Now().UTC().Format("2006-01-02"))
		log.Println("New Uuid was successfully created for ", msName)
	}

}

func main() {
	lambda.Start(Handler)
}
