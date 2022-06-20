// dlq
resource "aws_sqs_queue" "check_uuid_dlq" {
  name = "checkUUID_lambda_dlq"

  visibility_timeout_seconds = 3000

  tags = {
    Environment = "dev"
  }
}

