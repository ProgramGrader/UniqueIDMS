package dynamoDAO

import (
	"common"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
	"time"
)

// Get given UUID returns value uses getItem, poses problems if we want the range key for MsName to UUID
// should be a bool

func Get(clientConfig *dynamodb.Client, tableName string, msName string) (UUID string, date string, list []string) {

	getItemInput := &dynamodb.GetItemInput{
		TableName:      aws.String(tableName),
		ConsistentRead: aws.Bool(true),

		Key: map[string]types.AttributeValue{
			"ms-name": &types.AttributeValueMemberS{Value: msName},
		},
	}

	output, err := clientConfig.GetItem(context.TODO(), getItemInput)
	if err != nil {
		log.Fatalf("Failed to get item, %v", err)
	}

	if output.Item == nil {
		log.Println("Item not found: ")
	}

	err = attributevalue.Unmarshal(output.Item["UUID"], &UUID)
	err = attributevalue.Unmarshal(output.Item["date"], &date)

	err = attributevalue.Unmarshal(output.Item["list"], &list)
	if err != nil {
		log.Fatalf("unmarshal failed, %v", err)
	}

	return UUID, date, list

}

// Could make another get method that uses Query instead of getItem, allowing us to have a range key and not need to
// specify it

// Put creates/update a new entry in the Dynamodb
func Put(clientConfig *dynamodb.Client, tableName string, msName string, UUID string, date string) {

	var itemInput = dynamodb.PutItemInput{
		TableName: aws.String(tableName),

		Item: map[string]types.AttributeValue{
			"ms-name": &types.AttributeValueMemberS{Value: msName},
			"UUID":    &types.AttributeValueMemberS{Value: UUID},
			"date":    &types.AttributeValueMemberS{Value: date},
		},
	}

	_, err := clientConfig.PutItem(context.TODO(), &itemInput)
	if err != nil {
		log.Fatal("Error inserting value ", err)
	}
}

// Delete removes a item from the table given the key
func Delete(clientConfig *dynamodb.Client, tableName string, msName string) error {

	deleteInput := dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"ms-name": &types.AttributeValueMemberS{Value: msName},
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
					"ms-name": item["ms-name"],
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
func DeleteExpiredUUIDs(clientConfig *dynamodb.Client, tableName string, msName string) {

	loc, _ := time.LoadLocation("UTC")

	earliestAcceptedDate := time.Now().In(loc)
	latestAcceptedDate := time.Now().In(loc).Add(-720 * time.Hour) // 30 days from current time

	out, err := clientConfig.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(common.TableName),
		IndexName:              aws.String("ms-name"),
		KeyConditionExpression: aws.String("ms-name = :msName AND :edate BETWEEN :ldate"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":msName": &types.AttributeValueMemberS{Value: msName},
			":ldate":  &types.AttributeValueMemberS{Value: latestAcceptedDate.String()},
			":edate":  &types.AttributeValueMemberS{Value: earliestAcceptedDate.String()},
		},
	})
	if err != nil {
		print("Error querying expired dates")
		log.Fatal(err)
	}

	for _, item := range out.Items {
		_, err = clientConfig.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
			TableName: aws.String(tableName),
			Key: map[string]types.AttributeValue{
				"ms-name": item["ms-name"],
			},
		})
		if err != nil {
			print("Error Deleting Item")
			panic(err)
		}

	}

	fmt.Println(out.Items)

}
