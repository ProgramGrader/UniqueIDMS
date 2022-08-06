package dynamoDAO

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"src/common"
	"testing"
)

//TODO in tftest, you need to be able to test the 3 requests GET, PUT, POST

var localClientConfig, _ = common.CreateDynamoDbLocalClient()

var cfg, _ = config.LoadDefaultConfig(context.TODO())

var clientConfig = dynamodb.NewFromConfig(cfg)

type Item struct {
	msName string
	ULID   string
	date   string

	expectedULID string
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

		Put(clientConfig, test.msName, test.ULID, test.date)

		storedULID, storedDate, _, err := Get(clientConfig, test.msName)
		if err != nil {
			t.Fatalf("TestGetAndPut(), Failed. Get function threw a error")
		}

		if storedULID != test.ULID {
			t.Fatalf("TestGetAndPut(), Failed. Expected value was not found. Got %s expected %s", storedULID, test.expectedULID)
		} else if storedDate != test.expectedDate {
			t.Fatalf("TestGet(), Failed. Expected value was not found. Got %s expected %s", storedDate, test.expectedDate)
		}
	}

	// Getting a ULID that doesn't exist yet
	_, _, _, err := Get(clientConfig, "Microservice6")
	if err != nil {
		t.Fatalf("TestGetAndPut(), Get threw error on non existent value")
	}

	//DeleteAll(localClientConfig, common.TableName)
}

func TestDeleteExpiredULIDs(t *testing.T) {

	// Grabbing ms from list
	_, _, msList, err := Get(clientConfig, "list")
	if err != nil {
		t.Fatalf("TestGet(), Failed. DeleteExpiredULIDs get function threw a error")
	}

	for _, msName := range msList {
		deleteErr := DeleteExpiredULIDs(clientConfig, common.TableName, msName)
		if deleteErr != nil {
			t.Fatal("TestDeleteExpiredULIDs(), Failed to delete. Expected error to be nil.")
		}
	}

}
