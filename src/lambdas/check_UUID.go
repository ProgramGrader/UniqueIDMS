package lambdas

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
	"time"
)

// checks DB if uuid exist if not creates a new one

// TODO - Test handler function using SAM

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	cfg, _ := config.LoadDefaultConfig(context.TODO())
	dynamodbClient := dynamodb.NewFromConfig(cfg)

	fmt.Printf("event.HTTPMethod %v\n", request.HTTPMethod)
	fmt.Printf("event.QueryStringParameters %v\n", request.QueryStringParameters)
	fmt.Printf("event %v\n", request)

	msName := request.QueryStringParameters["ms_name"]

	var UUID string
	var statusCode int
	UUID, _ = dynamoDAO.Get(dynamodbClient, common.TableName, msName)
	if UUID == "" {

		dynamoDAO.Put(dynamodbClient, common.TableName, msName, uuid.New().String(), time.Now().Format("2006-05-10"))
		statusCode = 201
	} else {
		statusCode = 200
	}

	//url := "https://iuscsg.org"
	return events.APIGatewayProxyResponse{
		//Body:       fmt.Sprintf("{\"message\":\"Error occurred unmarshaling request: %v.\"}", url),
		StatusCode: statusCode,
		Body:       UUID,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
