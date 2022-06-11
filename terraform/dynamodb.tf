resource "aws_dynamodb_table" "MSUniqueID" {

  name     = "MSUniqueID"
  billing_mode = "PAY_PER_REQUEST"
  attribute {
    name = "MsName"
    type = "S"
  }
  hash_key = "MsName"
  //range_key = "date-created"

  attribute {
    name = "CreationDate"
    type = "S"
  }
  global_secondary_index {
    hash_key        ="CreationDate"  // partition key
    name            = "CreationDateIndex"
    projection_type = "ALL"
    range_key =  "MsName" // sort key
  }

}
// sort key can't be attached to the primary key or else the sort key will be required to get items.


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