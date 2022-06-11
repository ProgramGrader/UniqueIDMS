resource "aws_apigatewayv2_api" "unique_id_gw" {
  name          = "unique_id_gwa"
  protocol_type = "HTTP"
}

// Defining permissions so that API gateway has permissions to SendMessage to SQS queue

resource "aws_iam_role" "apigw-role" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "apigateway.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "template_file" "gateway_policy" {
  template = file("policies/api-gateway-permission.json")

  vars = {
    sqs_arn = aws_sqs_queue.sqs.arn
  }
}

resource "aws_iam_policy" "api-policy" {
  name   = "api-sqs-cloudwatch-policy"
  policy = data.template_file.gateway_policy.rendered
}

resource "aws_iam_role_policy_attachment" "api_exec_role" {
  policy_arn = aws_iam_policy.api-policy.arn
  role       = aws_iam_role.apigw-role.name
}

// creating api gateway
resource "aws_apigatewayv2_integration" "api" {
  api_id             = aws_apigatewayv2_api.unique_id_gw.id
  integration_type   = "AWS_PROXY"
  credentials_arn    = aws_iam_role.apigw-role.arn
  integration_subtype = "SQS-SendMessage"
  description = "Send SQS request"

  request_parameters = {
    QueueUrl = aws_sqs_queue.sqs.url
    // not sure what exactly to send to sqs
    // defaulting to querystring
    MessageBody = "$request.querystring.ms_name"

  }
  // need to add a request template in order to pass method Body and path in sqs
#  request_templates = {
#    "application/json" = <<EOF
#    Action=SendMessage&MessageBody={
#    "method": "$context.httpMethod",
#    "body-json" : $input.json('$'),
#    "queryParams": {
#      #foreach($param in $input.params().querystring.keySet())
#      "$param": "$util.escapeJavaScript($input.params().querystring.get($param))" #if($foreach.hasNext),#end
#    #end
#  },
#  "pathParams": {
#    #foreach($param in $input.params().path.keySet())
#    "$param": "$util.escapeJavaScript($input.params().path.get($param))" #if($foreach.hasNext),#end
#    #end
#  }
#}"
#EOF
#  }

  depends_on = [aws_iam_role_policy_attachment.api_exec_role]
}

resource "aws_apigatewayv2_route" "uniqueUUIDMS" {
  api_id    = aws_apigatewayv2_api.unique_id_gw.id
  route_key = "POST /{proxy+}"
  target = "integrations/${aws_apigatewayv2_integration.api.id}"
}

resource "aws_apigatewayv2_deployment" "api" {
  api_id        = aws_apigatewayv2_api.unique_id_gw.id
  lifecycle {
    create_before_destroy = true
  }


  depends_on = [aws_apigatewayv2_route.uniqueUUIDMS]
}


resource "aws_cloudwatch_log_group" "api-gw" {
  retention_in_days = 30
}

resource "aws_apigatewayv2_stage" "api-gw_stage" {
  api_id = aws_apigatewayv2_api.unique_id_gw.id
  name   = var.environment
  auto_deploy = true
  //PascalCase

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api-gw.arn

    format = jsonencode({
      requestId               = "$context.requestId"
      sourceIp                = "$context.identity.sourceIp"
      requestTime             = "$context.requestTime"
      protocol                = "$context.protocol"
      httpMethod              = "$context.httpMethod"
      resourcePath            = "$context.resourcePath"
      routeKey                = "$context.routeKey"
      status                  = "$context.status"
      responseLength          = "$context.responseLength"
      integrationErrorMessage = "$context.integrationErrorMessage"
    }
    )
  }
}