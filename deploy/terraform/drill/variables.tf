variable "region" {
  description = "AWS region for the drill environment."
  type        = string
  default     = "ap-southeast-1"
}

variable "name" {
  description = "Name prefix for resources and the SSM parameter path."
  type        = string
  default     = "vpsentinel-drill"

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{2,30}$", var.name))
    error_message = "name must be 3-31 lowercase letters, digits, or hyphens and start with a letter."
  }
}

variable "release_tag" {
  description = "VPSentinel GitHub release to deploy, e.g. v2026.10.10.1791234567. It must include the migrate job and installer assets."
  type        = string

  validation {
    condition     = can(regex("^v[0-9]{4}\\.[0-9]{2}\\.[0-9]{2}\\.[0-9]+$", var.release_tag))
    error_message = "release_tag must look like vYYYY.MM.DD.TIMESTAMP."
  }
}

variable "repository" {
  description = "GitHub owner/name that publishes the release and container images."
  type        = string
  default     = "z9wen/VPSentinel"
}

variable "agent_count" {
  description = "Number of disposable monitored nodes."
  type        = number
  default     = 2

  validation {
    condition     = var.agent_count >= 1 && var.agent_count <= 5
    error_message = "agent_count must be between 1 and 5."
  }
}

variable "control_plane_instance_type" {
  description = "Graviton (arm64) instance type for the control plane."
  type        = string
  default     = "t4g.small"

  validation {
    condition     = can(regex("^[a-z]+[0-9]+g[a-z]*\\.", var.control_plane_instance_type))
    error_message = "Use a Graviton (arm64) instance type such as t4g.small; the AMI is arm64."
  }
}

variable "agent_instance_type" {
  description = "Graviton (arm64) instance type for monitored nodes."
  type        = string
  default     = "t4g.nano"

  validation {
    condition     = can(regex("^[a-z]+[0-9]+g[a-z]*\\.", var.agent_instance_type))
    error_message = "Use a Graviton (arm64) instance type such as t4g.nano; the AMI is arm64."
  }
}

variable "vpc_cidr" {
  description = "CIDR block for the drill VPC."
  type        = string
  default     = "10.42.0.0/16"
}

variable "budget_alert_email" {
  description = "Email for AWS Budgets alerts. Leave empty to skip creating the budget."
  type        = string
  default     = ""
}

variable "budget_limit_usd" {
  description = "Monthly account-wide cost budget in USD."
  type        = number
  default     = 10
}
