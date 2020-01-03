terraform {
  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "ianlopshire-core"

    workspaces {
      name = "go-ssm-config"
    }
  }
}

provider "aws" {
  version = "~> 2.0"
  region  = "us-east-1"
}

resource "aws_ssm_parameter" "go_ssm_config_string_1" {
  name  = "/go-ssm-config/strings/s1"
  type  = "String"
  value = "string1"
}

resource "aws_ssm_parameter" "go_ssm_config_string_2" {
  name  = "/go-ssm-config/base/strings/s2"
  type  = "String"
  value = "string2"
}

resource "aws_ssm_parameter" "go_ssm_config_int_1" {
  name  = "/go-ssm-config/int/i1"
  type  = "String"
  value = "42"
}

resource "aws_ssm_parameter" "go_ssm_config_int_zero" {
  name  = "/go-ssm-config/int/i_zero"
  type  = "String"
  value = "0"
}

resource "aws_ssm_parameter" "go_ssm_config_bool_1" {
  name  = "/go-ssm-config/bool/b1"
  type  = "String"
  value = "true"
}

resource "aws_ssm_parameter" "go_ssm_config_bool_2" {
  name  = "/go-ssm-config/bool/b2"
  type  = "String"
  value = "false"
}

resource "aws_ssm_parameter" "go_ssm_config_float32_1" {
  name  = "/go-ssm-config/float32/f321"
  type  = "String"
  value = "42.42"
}

resource "aws_ssm_parameter" "go_ssm_config_float64_1" {
  name  = "/go-ssm-config/float64/f641"
  type  = "String"
  value = "42.42"
}