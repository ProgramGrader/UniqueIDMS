package test

import (
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDynamodbSuccess(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../",
	})

	terraform.InitAndApply(t, terraformOptions)
	terraform.Apply(t, terraformOptions)
	output := terraform.Output(t, terraformOptions, "dynamo_resource")
	assert.NotNil(t, output)

	terraform.Destroy(t, terraformOptions)
	// execute tests in dynamoDAO dir
	// check whether they actually exist in aws

}
