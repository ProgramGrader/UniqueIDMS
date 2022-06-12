#// grab this url to interact with the api_client
output "api_url"{ // url of the api gateway
  value = aws_apigatewayv2_stage.api-gw_stage.invoke_url
}

output "dynamo_resource" {
  value = "arn:aws:dynamodb:${var.primary_aws_region}:${aws_dynamodb_table.MSUniqueID.arn}"
}