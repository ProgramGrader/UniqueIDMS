package dynamoDAO

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
	"src/common"
	"time"
)

// Get given ULID returns value uses getItem, poses problems if we want the range key for MsName to ULID

func Get(clientConfig *dynamodb.Client, msName string) (ULID string, expirationDate string, list []string, err error) {
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

	}

	if len(query.Items) == 0 {
		log.Println("ULID for ", msName, " not found")
		return ULID, expirationDate, list, err
	}

	queryValues := query.Items[0]

	err = attributevalue.Unmarshal(queryValues["ULID"], &ULID)
	err = attributevalue.Unmarshal(queryValues["expirationDate"], &expirationDate)
	err = attributevalue.Unmarshal(queryValues["list"], &list)
	if err != nil {
		log.Println("Error unmarshalling data from query ")
	}

	return ULID, expirationDate, list, err

}

// Put creates/update a new entry in the Dynamodb
func Put(clientConfig *dynamodb.Client, msName string, ULID string, expirationDate string) {

	var itemInput = dynamodb.PutItemInput{
		TableName: aws.String(common.TableName),

		Item: map[string]types.AttributeValue{
			"ms-name":        &types.AttributeValueMemberS{Value: msName},
			"ULID":           &types.AttributeValueMemberS{Value: ULID},
			"expirationDate": &types.AttributeValueMemberS{Value: expirationDate},
		},
	}

	_, err := clientConfig.PutItem(context.TODO(), &itemInput)
	if err != nil {
		log.Fatal("Error inserting value ", err)
	}
}

// DeleteExpiredULIDs
// queries all ULIDs and creation expirationDates for a given ms and deletes items passed expiration date
func DeleteExpiredULIDs(clientConfig *dynamodb.Client, tableName string, msName string) error {

	loc, _ := time.LoadLocation("UTC")
	currentTime := time.Now().In(loc).Format("2006-01-02") // 30 * 24 = 720

	filter := "#expirationDate < :currentTime"
	out, err := clientConfig.Query(context.TODO(),
		&dynamodb.QueryInput{
			TableName:              aws.String(common.TableName),
			KeyConditionExpression: aws.String("#msName = :msName"),
			ExpressionAttributeNames: map[string]string{
				"#msName":         "ms-name",        // dynamodb does not like dashes, #msName is a alias for ms-name
				"#expirationDate": "expirationDate", // defining the other column we want to query
			},

			ExpressionAttributeValues: map[string]types.AttributeValue{
				":msName":      &types.AttributeValueMemberS{Value: msName}, // all expired ULIDs from the msName
				":currentTime": &types.AttributeValueMemberS{Value: currentTime},
			},
			FilterExpression: &filter,
		})
	if err != nil {
		print("Error querying expired expirationDates")
		log.Fatal(err)
	}

	for _, item := range out.Items {
		_, err = clientConfig.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
			TableName: aws.String(tableName),
			Key: map[string]types.AttributeValue{
				"ms-name": item["ms-name"],
				"ULID":    item["ULID"],
			},
		})
		if err != nil {
			print("Error Deleting Item")
			panic(err)
		}

	}
	return err
}
