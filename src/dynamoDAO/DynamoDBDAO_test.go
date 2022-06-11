package dynamoDAO

import (
	"common"
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"testing"
)

var localClientConfig, _ = common.CreateDynamoDbLocalClient()

var cfg, _ = config.LoadDefaultConfig(context.TODO())

var clientConfig = dynamodb.NewFromConfig(cfg)

type Item struct {
	msName string
	UUID   string
	date   string

	expectedUUID string
	expectedDate string
}

// yyyy-mm-dd
func TestPutAndGet(t *testing.T) {

	putAnGetItems := []Item{
		{"Microservice1", "cda6498a-235d-4f7e-ae19-661d41bc154c", "2022-01-22", "cda6498a-235d-4f7e-ae19-661d41bc154c", "2022-01-22"},
		{"Microservice2", "c8472cb6-da1c-48f5-8b61-72355fb647fa", "2022-03-22", "c8472cb6-da1c-48f5-8b61-72355fb647fa", "2022-03-22"},
		{"Microservice3", "332b1983-2ddd-4da9-aaf2-f2cf2b3e2009", "2022-05-22", "332b1983-2ddd-4da9-aaf2-f2cf2b3e2009", "2022-05-22"},
		{"Microservice4", "332b1983-2ddd-4da9-aaf2-f2cf2b3e2009", "2021-03-21", "332b1983-2ddd-4da9-aaf2-f2cf2b3e2009", "2021-03-21"},
	}

	for _, test := range putAnGetItems {

		Put(clientConfig, common.TableName, test.msName, test.UUID, test.date)

		storedUUID, storedDate := Get(clientConfig, common.TableName, test.msName)
		if storedUUID != test.UUID {
			t.Fatalf("TestGet(), Failed. Expected value was not found. Got %s expected %s", storedUUID, test.expectedUUID)
		} else if storedDate != test.expectedDate {
			t.Fatalf("TestGet(), Failed. Expected value was not found. Got %s expected %s", storedDate, test.expectedDate)
		}
	}

	//DeleteAll(localClientConfig, common.TableName)
}

func TestDelete(t *testing.T) {

	deleteItems := []Item{
		{"Microservice1", "cda6498a-235d-4f7e-ae19-661d41bc154c", "2022-01-20", "", ""},
		{"Microservice2", "c8472cb6-da1c-48f5-8b61-72355fb647fa", "2022-03-20", "", ""},
		{"Microservice3", "332b1983-2ddd-4da9-aaf2-f2cf2b3e2009", "2022-05-20", "", ""},
	}

	for _, test := range deleteItems {
		Put(clientConfig, common.TableName, test.msName, test.UUID, test.date)

		deleteErr := Delete(clientConfig, common.TableName, test.UUID)
		if deleteErr != nil {
			t.Fatal("TestDelete(), Failed to delete. Expected error to be nil.")
		}
	}
}

func TestDeleteExpiredUUIDs(t *testing.T) {
	DeleteExpiredUUIDs(clientConfig, common.TableName)
}
