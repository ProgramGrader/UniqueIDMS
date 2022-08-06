// dlq
resource "aws_sqs_queue" "check_ulid_dlq" {
  name = "checkULID_lambda_dlq"
  visibility_timeout_seconds = 3000
  tags = {
    Environment = "dev"
  }
}

