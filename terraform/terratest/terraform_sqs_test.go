package test

import (
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSQSSuccess(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../",
	})

	terraform.InitAndApply(t, terraformOptions)
	terraform.Apply(t, terraformOptions)
	checkUUIDOutput := terraform.Output(t, terraformOptions, "check_UUID_sqs")
	checkUUIDDlqOutput := terraform.Output(t, terraformOptions, "check_UUID_dlq")

	assert.NotNil(t, checkUUIDOutput)
	assert.NotNil(t, checkUUIDDlqOutput)

	terraform.Destroy(t, terraformOptions)
	// need to test DlQ response when lambda fails

}
