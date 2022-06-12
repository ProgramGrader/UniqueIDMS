//  boiler plate code
resource "aws_sqs_queue" "sqs" {
  name = "UniqueIdMS_sqs"
  delay_seconds = 90
  max_message_size = 2048
  message_retention_seconds = 86400
  receive_wait_time_seconds = 10

  //redrive_policy = "{\"deadLetterTargetArn\":\"${aws_sqs_queue.check_uuid_dlq.arn}\",\"maxReceiveCount\":4}"
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

