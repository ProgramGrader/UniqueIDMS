// Tools used to test this infrastructure locally: Localstacks, tflocal, and awslocal
// build localStacks: docker-compose up
// pip install terraform-local
// if the tflocal or awslocal commands aren't recognized try restarting your terminal

// TODO - Fix terraform vulnerabilities
// TODO - Test terraform using terragrunt

terraform {

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.15"
    }
  }

  backend "s3" {
    bucket = "tfstate-3ea6z45i"
    key    = "assignmentZip/key"
    region = "us-east-2"
    dynamodb_table = "app-state"
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
