package main

import (
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

	// url/env/v../Microservicename?UUID=UUID
	UUID, _, _, err := dynamoDAO.Get(dynamodbClient, msName)

	// Microservice provides UUID to insert into dynamodb
	if request.HTTPMethod == "PUT" {
		passedUUID := request.QueryStringParameters["UUID"]
		UUID, _, _, err = dynamoDAO.Get(dynamodbClient, msName)
		if UUID != "" {
			statusCode = 409
		} else {
			err = dynamoDAO.Put(dynamodbClient, msName, passedUUID, time.Now().UTC().Format("2006-01-02"))
			UUID = passedUUID
			statusCode = 201
		}
	}

	if request.HTTPMethod == "GET" && UUID == "" {
		UUID = uuid.New().String()
		// url/env/v../ms_name
		err = dynamoDAO.Put(dynamodbClient, msName, UUID, time.Now().UTC().Format("2006-01-02"))
		log.Println("New Uuid was successfully created for ", msName)
		UUID, _, _, err = dynamoDAO.Get(dynamodbClient, msName)
		statusCode = 201
	}

	// created in terraform (aws_cloudwatch_log_metric_filter) filters the 409 statuscode
	// which then triggers an alarm notifying of a collision
	log.Println("statusCode=", statusCode)

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       fmt.Sprintf("UUID=%v", UUID),
	}, err

}

func main() {
	lambda.Start(Handler)
}
