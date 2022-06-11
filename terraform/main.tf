// Tools used to test this infrastructure locally: Localstacks, tflocal, and awslocal
// build localStacks: docker-compose up
// pip install terraform-local
// if the tflocal or awslocal commands aren't recognized try restarting your terminal

// TODO - Fix terraform vulnerabilities
// TODO - Test terraform using terragrunt
// TODO - Test in production


// Local stacks does not support apigw v2 unless you have the pro version
terraform {

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.15"
    }
  }

}

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
