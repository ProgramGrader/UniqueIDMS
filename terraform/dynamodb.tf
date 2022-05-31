resource "aws_dynamodb_table" "MSUniqueID" {

  name     = "MSUniqueID"
  billing_mode = "PAY_PER_REQUEST"
  attribute {
    name = "UUID"
    type = "S"
  }
  hash_key = "UUID"
  range_key = "date-created"
  attribute {
    name = "date-created"
    type = "N"
  }
  global_secondary_index {
    hash_key        = "date-created"
    name            = "date-created"
    projection_type = "ALL"
  }

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