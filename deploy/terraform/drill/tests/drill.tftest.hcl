# Runs with mocked AWS responses, so `terraform test` needs no credentials and
# creates nothing. It checks the module's wiring and input validation.

mock_provider "aws" {
  mock_data "aws_ec2_instance_type_offerings" {
    defaults = {
      locations = ["ap-southeast-1b", "ap-southeast-1a"]
    }
  }

  mock_data "aws_ssm_parameter" {
    defaults = {
      insecure_value = "ami-0123456789abcdef0"
    }
  }

  mock_data "aws_caller_identity" {
    defaults = {
      account_id = "123456789012"
    }
  }

  mock_data "aws_partition" {
    defaults = {
      partition = "aws"
    }
  }

  mock_data "aws_iam_policy_document" {
    defaults = {
      json = "{\"Version\":\"2012-10-17\",\"Statement\":[]}"
    }
  }

  mock_resource "aws_instance" {
    defaults = {
      private_ip = "10.42.0.10"
    }
  }
}

variables {
  release_tag = "v2026.10.10.1791234567"
}

run "default_drill_environment" {
  command = apply

  assert {
    condition     = length(aws_instance.agent) == 2 && length(aws_ssm_parameter.agent_token) == 2
    error_message = "Expected two drill nodes, each with its own token parameter."
  }

  assert {
    condition     = aws_ssm_parameter.agent_token[1].name == "/vpsentinel-drill/agent-tokens/drill-2"
    error_message = "Agent token parameters must live under the environment prefix."
  }

  assert {
    condition     = aws_subnet.public.availability_zone == "ap-southeast-1a"
    error_message = "The subnet should use the first zone that offers both instance types."
  }

  assert {
    condition     = aws_instance.control_plane.metadata_options[0].http_tokens == "required" && aws_instance.control_plane.metadata_options[0].http_put_response_hop_limit == 1
    error_message = "The control plane must require IMDSv2 with a hop limit of 1."
  }

  assert {
    condition     = length(aws_budgets_budget.monthly) == 0
    error_message = "No budget should be created without an alert email."
  }

  assert {
    condition     = keys(output.agent_instance_ids) == ["drill-1", "drill-2"]
    error_message = "Outputs should map node names to instance IDs."
  }
}

run "budget_when_email_is_set" {
  command = apply

  variables {
    agent_count        = 3
    budget_alert_email = "ops@example.com"
  }

  assert {
    condition     = length(aws_budgets_budget.monthly) == 1 && length(aws_iam_role.agent) == 3
    error_message = "Expected a budget and one IAM role per drill node."
  }
}

run "rejects_x86_instance_types" {
  command = plan

  variables {
    agent_instance_type = "t3.micro"
  }

  expect_failures = [var.agent_instance_type]
}

run "rejects_malformed_release_tag" {
  command = plan

  variables {
    release_tag = "latest"
  }

  expect_failures = [var.release_tag]
}
