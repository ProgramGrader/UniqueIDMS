resource "aws_dynamodb_table" "MSUniqueID" {

  name     = "MSUniqueID"
  billing_mode = "PAY_PER_REQUEST"
  attribute {
    name = "ms-name"
    type = "S"
  }
  hash_key = "ms-name"
  //range_key = "date-created"

}
//
// sort key can't be attached to the primary key or else the sort key will be required to get items.

resource "aws_dynamodb_table_item" "ms_list" {
  hash_key   = aws_dynamodb_table.MSUniqueID.hash_key
  table_name = aws_dynamodb_table.MSUniqueID.name
  item = <<ITEM
    {
      "ms-name": {"S": "list"},
      "UUID": {"S": "list"},
      "date": {"S": "list"},
      "list": {"L": [{"S":"Microservice1"},{"S":"Microservice2"}]}
    }
    ITEM
}

