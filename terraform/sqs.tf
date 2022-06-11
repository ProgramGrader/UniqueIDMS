//  boiler plate code
resource "aws_sqs_queue" "sqs" {
  name = "UniqueIdMS_sqs"
  delay_seconds = 90
  max_message_size = 2048
  message_retention_seconds = 86400
  receive_wait_time_seconds = 10

  redrive_policy = "{\"deadLetterTargetArn\":\"${aws_sqs_queue.check_uuid_dlq.arn}\",\"maxReceiveCount\":4}"

  tags = {
    Environment = "dev"
  }

}

// dlq

resource "aws_sqs_queue" "check_uuid_dlq" {
  name = "checkUUID_lambda_dlq"

  visibility_timeout_seconds = 3000

  tags = {
    Environment = "dev"
  }
}

resource "aws_sns_topic" "check_uuid_lambda_failure" {
  name = "checkUUID_lambda_failure"
}

resource "aws_sns_topic_subscription" "check_uuid_lambda" {
  endpoint  = aws_lambda_function.check_UUID_lambda.arn
  protocol  = "lambda"
  topic_arn = aws_sns_topic.check_uuid_lambda_failure.arn
}

resource "aws_sns_topic" "check_uuid_lambda" {
  name = "checkUUID_lambda"
}

resource "aws_sns_topic_subscription" "check_uuid_lambda" {
  endpoint  = aws_lambda_function.check_UUID_lambda
  protocol  = "lambda"
  topic_arn = aws_sns_topic.check_uuid_lambda.arn
}