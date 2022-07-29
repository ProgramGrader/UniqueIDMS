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
        "Resource": [aws_sqs_queue.check_uuid_dlq.arn]
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

resource "aws_iam_role_policy_attachment" "attach_dynamodb_sqs_policy" {
  role       = aws_iam_role.lambda-role.name
  policy_arn = aws_iam_policy.dynamodb-sqs-policy.arn
}

resource "aws_iam_role_policy_attachment" "attach_dynamodb_rw_policy" {
  role       = aws_iam_role.lambda-role.name
  policy_arn = aws_iam_policy.readwrite-policy.arn

}

resource "aws_iam_role_policy_attachment" "attach_xray_write_access"{
  policy_arn = "arn:aws:iam::aws:policy/AWSXrayWriteOnlyAccess"
  role       = aws_iam_role.lambda-role.name
}

resource "aws_lambda_function" "check_UUID_lambda" {

  depends_on = [data.archive_file.check_UUID_lambda_zip,aws_iam_role_policy_attachment.attach_dynamodb_sqs_policy]

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

  // x-ray tracing
  tracing_config {
    mode = "Active"
  }

  // enabled enhanced monitoring
  layers = ["arn:aws:lambda:${var.primary_aws_region}:580247275435:layer:LambdaInsightsExtension:14"]

}

// Metric filter for filtering 409 statusCode errors, which imply a UUID Collision
resource "aws_cloudwatch_log_metric_filter" "UUID_Collision" {

  log_group_name = "/aws/lambda/check-uuid"
  name           = "IMSUUIDCollisions"
  pattern        = "[..., statusCode=409]"
  metric_transformation {
    name      = "UUIDCollisionFilter"
    namespace = "IMS"
    value     = "1"
  }
}

resource "aws_cloudwatch_metric_alarm" "UUID_Collision" {
  alarm_name          = "IMS UUID Collisions"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  period = "60"
  evaluation_periods  = 1
  metric_name = aws_cloudwatch_log_metric_filter.UUID_Collision.name
  namespace = "IMS"
  threshold = "1"
  alarm_description = "This alarm monitors whether a microservice has tried to put a UUID that already exists into the database"
  statistic = "Maximum"
}

// Allows lambda to enable enhanced monitoring
resource "aws_iam_role_policy_attachment" "insights_policy" {
  role       = aws_iam_role.lambda-role.name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchLambdaInsightsExecutionRolePolicy"
}

// adding permission to allow api gw to invoke the lambda
resource "aws_lambda_permission" "allow_apigw_to_trigger_lambda" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.check_UUID_lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.unique_id_gw.execution_arn}/*/*"

}

// scheduled uuid deleter Lambda config
resource "aws_lambda_function" "scheduled_UUID_deleter_lambda" {
  depends_on = [data.archive_file.schedule_UUID_deleter_lambda_zip]

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

  // enabling enhanced lambda monitoring
  layers = ["arn:aws:lambda:${var.primary_aws_region}:580247275435:layer:LambdaInsightsExtension:14"]
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

// TODO do more testing around the scheduled deletion lambda