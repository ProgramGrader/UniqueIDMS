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
  integration_type   = "AWS"
  integration_method = "POST"
  credentials_arn    = aws_iam_role.apigw-role.arn

  // need to add a request template in order to pass method Body and path in sqs
  request_templates = {
    "application/json" = <<EOF
    Action=SendMessage&MessageBody={
    "method": "$context.httpMethod",
    "body-json" : $input.json('$'),
    "queryParams": {
      #foreach($param in $input.params().querystring.keySet())
      "$param": "$util.escapeJavaScript($input.params().querystring.get($param))" #if($foreach.hasNext),#end
    #end
  },
  "pathParams": {
    #foreach($param in $input.params().path.keySet())
    "$param": "$util.escapeJavaScript($input.params().path.get($param))" #if($foreach.hasNext),#end
    #end
  }
}"
EOF
  }

  depends_on = [aws_iam_role_policy_attachment.api_exec_role]
}

resource "aws_apigatewayv2_integration_response" "http200" {
  api_id                   = aws_apigatewayv2_api.unique_id_gw.id
  integration_id           = aws_apigatewayv2_integration.api.id
  integration_response_key = "/200/"
}

resource "aws_apigatewayv2_deployment" "api" {
  api_id        = aws_apigatewayv2_api.unique_id_gw.id
  lifecycle {
    create_before_destroy = true
  }
}