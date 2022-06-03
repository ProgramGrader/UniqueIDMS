package dynamoDAO

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
	"time"
)

// Get given UUID returns value
func Get(clientConfig *dynamodb.Client, tableName string, UUID string) (msName string, date string) {

	getItemInput := &dynamodb.GetItemInput{
		TableName:      aws.String(tableName),
		ConsistentRead: aws.Bool(true),

		Key: map[string]types.AttributeValue{
			"UUID": &types.AttributeValueMemberS{Value: UUID},
		},
	}

	output, err := clientConfig.GetItem(context.TODO(), getItemInput)
	if err != nil {
		log.Fatalf("Failed to get item, %v", err)
	}

	if output.Item == nil {
		log.Fatal("Item not found: ", UUID)
	}

	err = attributevalue.Unmarshal(output.Item["date"], &date)
	if err != nil {
		log.Fatalf("unmarshal failed, %v", err)
	}

	return msName, date

}

// Put creates/update a new entry in the Dynamodb
func Put(clientConfig *dynamodb.Client, tableName string, msName string, UUID string, date string) {

	var itemInput = dynamodb.PutItemInput{
		TableName: aws.String(tableName),

		Item: map[string]types.AttributeValue{
			"msName":       &types.AttributeValueMemberS{Value: msName},
			"UUID":         &types.AttributeValueMemberS{Value: UUID},
			"date-created": &types.AttributeValueMemberS{Value: date},
		},
	}

	_, err := clientConfig.PutItem(context.TODO(), &itemInput)
	if err != nil {
		log.Fatal("Error inserting value ", err)
	}
}

// Delete removes a item from the table given the key
func Delete(clientConfig *dynamodb.Client, tableName string, UUID string) error {

	deleteInput := dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"UUID": &types.AttributeValueMemberS{Value: UUID},
		},
	}

	_, err := clientConfig.DeleteItem(context.TODO(), &deleteInput)
	if err != nil {
		panic(err)
	}

	return err
}

func DeleteAll(clientConfig *dynamodb.Client, tableName string) {
	scan := dynamodb.NewScanPaginator(clientConfig, &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})

	for scan.HasMorePages() {
		out, err := scan.NextPage(context.TODO())
		if err != nil {
			print("Page error")
			panic(err)
		}

		for _, item := range out.Items {
			_, err = clientConfig.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
				TableName: aws.String(tableName),
				Key: map[string]types.AttributeValue{
					"UUID": item["UUID"],
				},
			})
			if err != nil {
				print("Error Deleting Item")
				panic(err)
			}

		}
	}
}

// DeleteExpiredUUIDs Expired UUIDs have persisted for 30 days or longer
func DeleteExpiredUUIDs(clientConfig *dynamodb.Client, tableName string) {

	loc, _ := time.LoadLocation("UTC")

	earliestAcceptedDate := time.Now().In(loc)
	latestAcceptedDate := time.Now().In(loc).Add(-720 * time.Hour) // 30 days from current time

	scan := dynamodb.NewScanPaginator(clientConfig, &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})

	for scan.HasMorePages() {
		out, err := scan.NextPage(context.TODO())
		if err != nil {
			print("Page error")
			panic(err)
		}

		for _, item := range out.Items {
			var date string
			err := attributevalue.Unmarshal(item["date-created"], &date)
			if err != nil {
				print("Error UnMarshaling date ", date)
				return
			}

			dateTime, err := time.Parse("2006-01-02", date)

			if dateTime.After(latestAcceptedDate) || dateTime.Before(earliestAcceptedDate) || dateTime.Equal(latestAcceptedDate) || dateTime.Equal(earliestAcceptedDate) {
				_, err = clientConfig.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
					TableName: aws.String(tableName),
					Key: map[string]types.AttributeValue{
						"UUID": item["UUID"],
					},
				})
				if err != nil {
					print("Error Deleting Item")
					panic(err)
				}
			}
		}
	}

}
