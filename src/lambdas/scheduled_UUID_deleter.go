package main

import (
	"common"
	"context"
	"dynamoDAO"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"log"
)

// Finds any UUID that is more than 30 days old then deletes it
// TODO - Test handler function using SAM

func Handler() {

	cfg, _ := config.LoadDefaultConfig(context.TODO())
	dynamodbClient := dynamodb.NewFromConfig(cfg)

	_, _, msList := dynamoDAO.Get(dynamodbClient, common.TableName, "list")
	// return error if something happens and print value
	for _, msName := range msList {
		err := dynamoDAO.DeleteExpiredUUIDs(dynamodbClient, common.TableName, msName)
		if err != nil {
			log.Println("Error Deleting Expired UUIDs")
		}
	}
}

func main() {
	lambda.Start(Handler)
}
