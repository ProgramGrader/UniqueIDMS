package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/oklog/ulid/v2"
	"log"
	"math/rand"
	"src/dynamoDAO"
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
	ULID := ""
	statusCode := 200

	// url/env/v../Microservicename?UUID=UUID
	ULID, _, _, err := dynamoDAO.Get(dynamodbClient, msName)

	// Microservice provides UUID to insert into dynamodb
	if request.HTTPMethod == "PUT" {
		passedULID := request.QueryStringParameters["ULID"]
		ULID, _, _, err = dynamoDAO.Get(dynamodbClient, msName)
		if ULID != "" {
			statusCode = 409
		} else {
			dynamoDAO.Put(dynamodbClient, msName, passedULID, time.Now().UTC().Add(+720).Format("2006-01-02T15:04:05-0700")) //30days * 24hrs = 720
			ULID = passedULID
			statusCode = 201
		}
	}

	if request.HTTPMethod == "GET" && ULID == "" {
		entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
		ms := ulid.Timestamp(time.Now())
		newULID, ulidErr := ulid.New(ms, entropy)
		if ulidErr != nil {
			log.Fatal("Failed to create a new ULID for: ", msName)
		}

		ULID = newULID.String()
		// url/env/v../ms_name
		loc, _ := time.LoadLocation("UTC")
		// default expiration date 30 days, 30*24=720

		dynamoDAO.Put(dynamodbClient, msName, ULID, time.Now().In(loc).Add(+720*time.Hour).Format("2006-01-02T15:04:05-0700"))
		log.Println("New Uuid was successfully created for ", msName)
		ULID, _, _, err = dynamoDAO.Get(dynamodbClient, msName)
		statusCode = 201
	}

	type Microservice struct {
		ExpirationTime string `json:"expirationTime"`
	}

	// Microservice provides body that defines expiration date
	if request.HTTPMethod == "POST" {
		ULID, _, _, err = dynamoDAO.Get(dynamodbClient, msName)
		if ULID != "" {
			statusCode = 409
		} else if request.Body != "" {

			body := fmt.Sprintf(`%s`, request.Body)
			expirationDateJson := []byte(body)
			ms := Microservice{}
			err = json.Unmarshal(expirationDateJson, &ms)
			if err != nil {
				fmt.Println(ms)
				log.Fatal("Failed to Unmarshal expiredTime json from api request body", err)
			}

			if ms.ExpirationTime == "" {
				log.Fatal("Inputted UserBody does not contain a accepted value in its contents")
			}
			dynamoDAO.Put(dynamodbClient, msName, ULID, ms.ExpirationTime)
			statusCode = 201
		} else {
			loc, _ := time.LoadLocation("UTC")
			dynamoDAO.Put(dynamodbClient, msName, ULID, time.Now().In(loc).Add(+720*time.Hour).Format("2006-01-02T15:04:05-0700"))
			statusCode = 201
		}

	}

	// in terraform (aws_cloudwatch_log_metric_filter) the statuscodes from lambda logs are filtered;
	// an alarm is triggered notifying of a collision if status code is 409
	log.Println("statusCode=", statusCode)

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       fmt.Sprintf("ULID=%v", ULID),
	}, err

}

func main() {
	lambda.Start(Handler)
}
