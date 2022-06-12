// compile code into binary
resource "null_resource" "compile_check_UUID_binary" {
  triggers = {
    build_number = timestamp()

  }
  provisioner "local-exec" {
    command = "GOOS=linux GOARCH=amd64 go build -ldflags '-w' -o  ../src/lambdas/check_UUID  ../src/lambdas/check_UUID.go"
  }
}

resource "null_resource" "compiled_scheduled_UUID_deleter_binary" {
  triggers = {
    build_number = timestamp()
  }

  provisioner "local-exec" {
    command = "GOOS=linux GOARCH=amd64 go build -ldflags '-w' -o  ../src/lambdas/scheduled_UUID_deleter  ../src/lambdas/scheduled_UUID_deleter.go"
  }
}

// zipping code
data "archive_file" "check_UUID_lambda_zip" {
  source_file = "../src/lambdas/check_UUID"
  type        = "zip"
  output_path = "check_UUID.zip"
  depends_on  = [null_resource.compile_check_UUID_binary]
}

data "archive_file" "schedule_UUID_deleter_lambda_zip" {
  source_file = "../src/lambdas/scheduled_UUID_deleter"
  type        = "zip"
  output_path = "scheduled_UUID_deleter.zip"
  depends_on  = [null_resource.compiled_scheduled_UUID_deleter_binary]
}

resource "aws_iam_role" "lambda-role" {
  assume_role_policy = jsonencode(
    {
      Version = "2012-10-17"
      Statement = [{
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
      ]
    })
}

// Allows necessary lambda and dynamodb permissions
resource "aws_iam_policy" "dynamodb-sqs-policy" {
  policy = jsonencode(
    {
      "Version": "2012-10-17",
      "Statement": [{
        "Sid": "ReadWriteTable",
        "Effect": "Allow",
        "Action": ["dynamodb:GetItem",
          "dynamodb:PutItem"],
        "Resource": "arn:aws:dynamodb:${var.primary_aws_region}:${aws_dynamodb_table.MSUniqueID.arn}"
      },
        {
          "Action": ["sqs:DeleteMessage",
            "sqs:ReceiveMessage",
            "sqs:SendMessage",
            "sqs:GetQueueAttributes"]
          "Resource": [aws_sqs_queue.sqs.arn,aws_sqs_queue.check_uuid_dlq.arn]
          "Effect": "Allow"
        },
        {
          "Action": [
            "logs:CreateLogGroup",
            "logs:CreateLogStream",
            "logs:PutLogEvents"
          ],
          "Resource": "arn:aws:logs:*:*:*",
          "Effect": "Allow"
        }
      ]
    })
}


resource "aws_iam_role_policy_attachment" "attach_dynamodb_sqs_policy" {
  role       = aws_iam_role.lambda-role.name
  policy_arn = aws_iam_policy.dynamodb-sqs-policy.arn
}

// Gives readOnly permission for dynamo
resource "aws_iam_policy" "readwrite-policy" {
  policy = jsonencode({
    "Version": "2012-10-17",
    "Statement": [{
      "Action": [
        "dynamodb:BatchGetItem",
        "dynamodb:Describe*",
        "dynamodb:List*",
        "dynamodb:GetItem",
        "dynamodb:PutItem",
        "dynamodb:Query",
        "dynamodb:Scan",
        "dynamodb:PartiQLSelect"

      ],
      "Effect": "Allow",
      "Resource": "*"
    },
      {
        "Action": "cloudwatch:GetInsightRuleReport",
        "Effect": "Allow",
        "Resource": "arn:aws:cloudwatch:*:*:insight-rule/DynamoDBContributorInsights*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "attach_dynamodb_rw_policy" {
  role       = aws_iam_role.lambda-role.name
  policy_arn = aws_iam_policy.readwrite-policy.arn

}

resource "aws_lambda_function" "check_UUID_lambda" {
  function_name = "check-uuid"
  filename = data.archive_file.check_UUID_lambda_zip.output_path
  source_code_hash = data.archive_file.check_UUID_lambda_zip.output_base64sha256
  handler = "check_UUID"
  role          = aws_iam_role.lambda-role.arn
  runtime = "go1.x"
  timeout = 5
  memory_size = 128

  dead_letter_config {
    target_arn = aws_sqs_queue.check_uuid_dlq.arn
  }

  tracing_config {
    mode = "Active"
  }
}
// adding permission to allow sqs to invoke the lambda
resource "aws_lambda_permission" "allow_sqs_to_trigger_lambda" {
  statement_id  = "AllowExecutionFromSQS"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.check_UUID_lambda.function_name
  principal     = "sqs.amazonaws.com"
  source_arn    = aws_sqs_queue.sqs.arn
}


// this is what connects the lambda event source to sqs
resource "aws_lambda_event_source_mapping" "event_source_mapping" {
  function_name = "check-uuid"
  event_source_arn = aws_sqs_queue.sqs.arn
  enabled = true  // allows the immediate sending of events to lambda
  batch_size = 5
}


// scheduled uuid deleter Lambda config

resource "aws_lambda_function" "scheduled_UUID_deleter_lambda" {
  function_name = "scheduled-uuid-deleter"
  filename = data.archive_file.schedule_UUID_deleter_lambda_zip.output_path
  source_code_hash = data.archive_file.schedule_UUID_deleter_lambda_zip.output_base64sha256
  handler = "scheduled_UUID_deleter"
  role          = aws_iam_role.lambda-role.arn
  runtime = "go1.x"
  timeout = 5
  memory_size = 128

  tracing_config {
    mode = "Active"
  }
}

resource "aws_cloudwatch_event_rule" "every_day" {
  name = "every_day"
  description = "Kicks of event every day"
  schedule_expression = "rate(24 hours)"
}

resource "aws_cloudwatch_event_target" "scheduled_UUID_deleter" {
  arn  = aws_lambda_function.scheduled_UUID_deleter_lambda.arn
  rule = aws_cloudwatch_event_rule.every_day.name
  target_id = "scheduled_UUID_deleter"

}

resource "aws_lambda_permission" "allow_cloudwatch_to_call_SUUID_deleter" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.scheduled_UUID_deleter_lambda.function_name
  principal     = "events.amazonaws.com"
  source_arn = aws_cloudwatch_event_rule.every_day.arn
  statement_id = "AllowExecutionFromCloudWatch"
}
