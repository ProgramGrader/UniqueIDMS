// Tools used to test this infrastructure locally: Localstacks, tflocal, and awslocal
// build localStacks: docker-compose up
// pip install terraform-local
// if the tflocal or awslocal commands aren't recognized try restarting your terminal

// TODO - Fix terraform vulnerabilities

// Local stacks does not support apigw v2 unless you have the pro version

// TODO - add sms and email alert to uuid collision alarm
locals {
  shared_tags = {
    Terraform = "true"
    Project = "UniqueIDMiS"
  }
}

provider "aws" {
  alias  = "primary"
  region = var.primary_aws_region

  default_tags {
    tags = local.shared_tags
  }
}

provider "aws" {
  alias  = "secondary"
  region = var.secondary_aws_region

  default_tags {
    tags = local.shared_tags
  }
}
