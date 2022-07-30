
// gives lambda permission to perform the below operations
resource "aws_iam_policy" "sqs-policy" {
  policy = jsonencode(
    {
      "Version": "2012-10-17",
      "Statement": [
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

#################################
## CHECK UUID LAMBDA CREATION ###
#################################

module "check_UUID" {
  source = "git::https://github.com/ProgramGrader/TerraformModules.git//modules//aws-lambda-apigwv2-creation"
  connect_to_api = true
  lambda_name = "check_UUID"
  lambda_path = "../src/lambdas/check_UUID.go"
  lambda_role_name = "IMSCheckUUIDLambdaRole"

  api_id = aws_apigatewayv2_api.unique_id_gw.id
  api_execution_arn = aws_apigatewayv2_api.unique_id_gw.execution_arn
  api_route_key = "ANY /{proxy+}"

  add_policies = 1 // access to dynamodbTable policy
  primary_aws_region = var.primary_aws_region
}

resource "aws_iam_role_policy_attachment" "attach_dynamodb_sqs_policy" {
  role       = module.check_UUID.lambda_role
  policy_arn = aws_iam_policy.sqs-policy.arn
}

resource "aws_lambda_function_event_invoke_config" "check_UUID" {
  function_name = module.check_UUID.lambda_name

  destination_config {
    on_failure {
      destination = aws_sqs_queue.check_uuid_dlq.arn
    }
  }
}

####################
## ALARM CREATION ##
####################

resource "aws_cloudwatch_log_metric_filter" "UUID_Collision" {

  log_group_name = "/aws/lambda/check_UUID"
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
  metric_name = aws_cloudwatch_log_metric_filter.UUID_Collision.metric_transformation[0].name
  namespace = "IMS"
  threshold = "1"
  alarm_description = "This alarm monitors whether a microservice has tried to put a UUID that already exists into the database"
  statistic = "Maximum"
}


####################################
## SCHEDULED UUID DELETION LAMBDA ##
####################################

module "scheduled_UUID_deleter_lambda" {
  source = "git::https://github.com/ProgramGrader/TerraformModules.git//modules//aws-lambda-apigwv2-creation"
  connect_to_api = false
  lambda_name = "scheduled_UUID_deleter"
  lambda_path = "../src/lambdas/scheduled_UUID_deleter.go"
  lambda_role_name = "IMSScheduledDeleterLambdaRole"
  primary_aws_region = var.primary_aws_region
}

resource "aws_cloudwatch_event_rule" "every_day" {
  name = "every_day"
  description = "Kicks of event every day"
  schedule_expression = "rate(24 hours)"
}

resource "aws_cloudwatch_event_target" "scheduled_UUID_deleter" {
  arn  = module.scheduled_UUID_deleter_lambda.lambda_arn
  rule = aws_cloudwatch_event_rule.every_day.name
  target_id = "scheduled_UUID_deleter"
}

resource "aws_lambda_permission" "allow_cloudwatch_to_call_scheduled_deleter" {
  action        = "lambda:InvokeFunction"
  function_name = module.scheduled_UUID_deleter_lambda.lambda_name
  principal     = "events.amazonaws.com"
  source_arn = aws_cloudwatch_event_rule.every_day.arn
  statement_id = "AllowExecutionFromCloudWatch"
}

// TODO do more testing around the scheduled deletion lambda