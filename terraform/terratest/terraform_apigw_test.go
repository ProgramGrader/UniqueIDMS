package test

// TODO make a test to check runtime should be less 10milisec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"testing"
)

func TestAPIGWSuccess(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../",
	})

	//defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)
	terraform.Apply(t, terraformOptions)
	output := terraform.Output(t, terraformOptions, "api_url")

	assert.NotNil(t, output)

	terraform.Destroy(t, terraformOptions)

}

func TestSendPostAPIGWSuccess(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../",
	})

	terraform.InitAndApply(t, terraformOptions)
	terraform.Apply(t, terraformOptions)
	output := terraform.Output(t, terraformOptions, "api_url")

	values := "{Microservice1}"
	json_data, err := json.Marshal(values)

	resp, err := http.Post(output+"/", "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		log.Fatal(err)
	}

	var postResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&postResp)
	if err != nil {
		log.Println("Failed to Decode response", err)
	}
	fmt.Println(postResp["json"])

	terraform.Destroy(t, terraformOptions)
}
