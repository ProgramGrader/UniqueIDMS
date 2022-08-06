package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"log"
	"src/common"
	"src/dynamoDAO"
)

// Finds any UUID that is more than 30 days old then deletes it
// TODO - Test handler function using SAM

func Handler() {

	cfg, _ := config.LoadDefaultConfig(context.TODO())
	dynamodbClient := dynamodb.NewFromConfig(cfg)

	_, _, msList, _ := dynamoDAO.Get(dynamodbClient, "list")
	// return error if something happens and print value
	for _, msName := range msList {
		err := dynamoDAO.DeleteExpiredULIDs(dynamodbClient, common.TableName, msName)
		if err != nil {
			log.Println("Error Deleting Expired ULIDs")
		}
	}
}

func main() {
	lambda.Start(Handler)
}
