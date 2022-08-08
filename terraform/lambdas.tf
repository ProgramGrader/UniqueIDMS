
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
        "Resource": [aws_sqs_queue.check_ulid_dlq.arn]
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
## CHECK ULID LAMBDA CREATION ###
#################################

module "check_ULID" {
  source = "git::https://github.com/ProgramGrader/TerraformModules.git//modules//aws-lambda-apigwv2-creation"

  lambda_name = "check_ULID"
  lambda_go_file = "check_ULID.go"
  lambda_directory = "../src/lambdas/check_ULID"
  lambda_role_name = "IMSCheckULIDLambda"

  connect_to_api = true
  api_id = aws_apigatewayv2_api.unique_id_gw.id
  api_execution_arn = aws_apigatewayv2_api.unique_id_gw.execution_arn
  api_route_key = "ANY /{proxy+}"

  add_policies = 1 // access to dynamodbTable policy
  primary_aws_region = var.primary_aws_region
}

resource "aws_iam_role_policy_attachment" "attach_dynamodb_sqs_policy" {
  role       = module.check_ULID.lambda_role
  policy_arn = aws_iam_policy.sqs-policy.arn
}

resource "aws_lambda_function_event_invoke_config" "check_ULID" {
  depends_on = [module.check_ULID]
  function_name = module.check_ULID.lambda_name

  destination_config {
    on_failure {
      destination = aws_sqs_queue.check_ulid_dlq.arn
    }
  }
}

####################
## ALARM CREATION ##
####################

resource "aws_cloudwatch_log_metric_filter" "ULID_Collision" {
 depends_on = [module.check_ULID]
  log_group_name = "/aws/lambda/check_ULID"
  name           = "IMS ULID Collisions"
  pattern        = "[..., statusCode=409]"
  metric_transformation {
    name      = "ULIDCollisionFilter"
    namespace = "IMS"
    value     = "1"
  }
}


resource "aws_cloudwatch_metric_alarm" "ULID_Collision" {
  alarm_name          = "IMSULIDCollisions"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  period = "60"
  evaluation_periods  = 1
  metric_name = aws_cloudwatch_log_metric_filter.ULID_Collision.metric_transformation[0].name
  namespace = "IMS"
  threshold = "1"
  alarm_description = "This alarm monitors whether a microservice has tried to put a ULID that already exists into the database"
  statistic = "Maximum"
}


####################################
## SCHEDULED UUID DELETION LAMBDA ##
####################################

module "scheduled_ULID_deleter_lambda" {
  depends_on = [module.check_ULID]
  source = "git::https://github.com/ProgramGrader/TerraformModules.git//modules//aws-lambda-apigwv2-creation"

  lambda_name = "scheduled_ULID_deleter"
  lambda_go_file = "scheduled_ULID_deleter.go"
  lambda_directory = "../src/lambdas/scheduled_ULID_deleter"
  lambda_role_name = "IMSScheduledDeleterLambda"

  connect_to_api = false
  primary_aws_region = var.primary_aws_region
}

resource "aws_cloudwatch_event_rule" "every_day" {
  name = "every_day"
  description = "Kicks of event every day"
  schedule_expression = "rate(24 hours)"
}

resource "aws_cloudwatch_event_target" "scheduled_ULID_deleter" {
  arn  = module.scheduled_ULID_deleter_lambda.lambda_arn
  rule = aws_cloudwatch_event_rule.every_day.name
  target_id = "scheduled_ULID_deleter"
}

resource "aws_lambda_permission" "allow_cloudwatch_to_call_scheduled_deleter" {
  action        = "lambda:InvokeFunction"
  function_name = module.scheduled_ULID_deleter_lambda.lambda_name
  principal     = "events.amazonaws.com"
  source_arn = aws_cloudwatch_event_rule.every_day.arn
  statement_id = "AllowExecutionFromCloudWatch"
}

// TODO do more testing around the scheduled deletion lambda