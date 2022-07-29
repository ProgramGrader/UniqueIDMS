package dynamoDAO

import (
	"common"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
	"time"
)

// Get given UUID returns value uses getItem, poses problems if we want the range key for MsName to UUID

func Get(clientConfig *dynamodb.Client, msName string) (UUID string, date string, list []string, err error) {
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(common.TableName),
		KeyConditionExpression: aws.String("#msName = :msName"),
		ExpressionAttributeNames: map[string]string{
			"#msName": "ms-name",
		},

		ExpressionAttributeValues: map[string]types.AttributeValue{
			":msName": &types.AttributeValueMemberS{Value: msName},
		},
	}

	query, err := clientConfig.Query(context.TODO(), queryInput)
	if err != nil {
		log.Fatal("Get() Failed to query values:", err)
		return "", "", nil, err
	}

	if len(query.Items) == 0 {
		log.Println("UUID for", msName, "not found")
		return "", "", nil, err
	}

	queryValues := query.Items[0]

	err = attributevalue.Unmarshal(queryValues["UUID"], &UUID)
	err = attributevalue.Unmarshal(queryValues["date"], &date)
	err = attributevalue.Unmarshal(queryValues["list"], &list)
	if err != nil {
		log.Println("Error unmarshalling data from query ")
	}

	return UUID, date, list, err

}

// Put creates/update a new entry in the Dynamodb
func Put(clientConfig *dynamodb.Client, msName string, UUID string, date string) error {

	var itemInput = dynamodb.PutItemInput{
		TableName: aws.String(common.TableName),

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
	return err
}

// DeleteExpiredUUIDs
// queries all UUIDs and creation dates for a given ms and deletes items that were created between the
// ranges 30 & 180 days before current date
func DeleteExpiredUUIDs(clientConfig *dynamodb.Client, tableName string, msName string) error {

	loc, _ := time.LoadLocation("UTC")
	sixMonthsAgo := time.Now().In(loc).Add(-4320 * time.Hour).Format("2006-01-02") // 30days * 12months = 180 * 24 = 4320
	monthAgo := time.Now().In(loc).Add(-720 * time.Hour).Format("2006-01-02")

	// in english: filter dates existing in dynamodb between 6 months ago and a month ago from current time
	filter := "#date BETWEEN :ldate AND :edate"
	out, err := clientConfig.Query(context.TODO(),
		&dynamodb.QueryInput{
			TableName:              aws.String(common.TableName),
			KeyConditionExpression: aws.String("#msName = :msName"),
			ExpressionAttributeNames: map[string]string{
				"#msName": "ms-name",
				"#date":   "date", // dynamodb does not like dashes
			},

			ExpressionAttributeValues: map[string]types.AttributeValue{
				":msName": &types.AttributeValueMemberS{Value: msName},
				":ldate":  &types.AttributeValueMemberS{Value: sixMonthsAgo},
				":edate":  &types.AttributeValueMemberS{Value: monthAgo},
			},
			FilterExpression: &filter,
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
				"UUID":    item["UUID"],
			},
		})
		if err != nil {
			print("Error Deleting Item")
			panic(err)
		}

	}
	return err
}
