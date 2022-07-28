package main

import (
	"common"
	"context"
	"dynamoDAO"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	"log"
	"time"
)

//TODO - Test handler function using SAM

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	cfg, _ := config.LoadDefaultConfig(context.TODO())
	dynamodbClient := dynamodb.NewFromConfig(cfg)
	log.Println("Headers: ", request.Headers)
	log.Println("Path: ", request.Path)
	log.Println("PathParams: ", request.PathParameters)
	log.Println("RequestCont: ", request.RequestContext)

	msName := request.PathParameters["proxy"]
	UUID := ""
	statusCode := 200

	if request.HTTPMethod == "GET" {
		// url/env/v../Microservicename?UUID=UUID
		UUID, _, _ = dynamoDAO.Get(dynamodbClient, common.TableName, msName)
		log.Println("UUID for ", msName, ":", UUID)

	}

	if UUID == "" {
		UUID = uuid.New().String()
		// url/env/v../ms_name

		dynamoDAO.Put(dynamodbClient, common.TableName, msName, UUID, time.Now().UTC().Format("2006-01-02"))
		log.Println("New Uuid was successfully created for ", msName)
		UUID, _, _ = dynamoDAO.Get(dynamodbClient, common.TableName, msName)
		statusCode = 201
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       fmt.Sprintf("UUID=%v", UUID),
	}, nil

}

func main() {
	lambda.Start(Handler)
}
