resource "aws_apigatewayv2_api" "unique_id_gw" {
  name          = "unique_id_gw"
  protocol_type = "HTTP"
}

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
}

resource "aws_iam_policy" "api-policy" {
  policy = data.template_file.gateway_policy.rendered
}

resource "aws_iam_role_policy_attachment" "api_exec_role" {
  policy_arn = aws_iam_policy.api-policy.arn
  role       = aws_iam_role.apigw-role.name
}

resource "aws_apigatewayv2_deployment" "api" {
  depends_on = [module.check_ULID]
  api_id        = aws_apigatewayv2_api.unique_id_gw.id
  lifecycle {
    create_before_destroy = true
  }
}


resource "aws_cloudwatch_log_group" "api-gw" {
  retention_in_days = 30
}

resource "aws_apigatewayv2_stage" "api-gw_stage" {
  api_id = aws_apigatewayv2_api.unique_id_gw.id
  name   = var.environment
  auto_deploy = true


#  access_log_settings {
#    destination_arn = aws_cloudwatch_log_group.api-gw.arn
#
#    format = jsonencode({
#      requestId               = "$context.requestId"
#      sourceIp                = "$context.identity.sourceIp"
#      requestTime             = "$context.requestTime"
#      protocol                = "$context.protocol"
#      httpMethod              = "$context.httpMethod"
#      resourcePath            = "$context.resourcePath"
#      routeKey                = "$context.routeKey"
#      status                  = "$context.status"
#      responseLength          = "$context.responseLength"
#      integrationErrorMessage = "$context.integrationErrorMessage"
#    }
    #)
 # }
}