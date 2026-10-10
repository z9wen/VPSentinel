# Terraform owns the parameters so `terraform destroy` removes every secret.
# The control plane overwrites the placeholders during bootstrap; ignoring the
# value keeps later plans from resetting them. Terraform still reads the real
# values into local state on refresh, which is acceptable only because this
# environment is disposable and its state never leaves the operator's machine.

locals {
  parameter_prefix    = "/${var.name}"
  pending_placeholder = "pending"
  node_names          = [for index in range(var.agent_count) : "drill-${index + 1}"]
}

resource "aws_ssm_parameter" "admin_password" {
  name        = "${local.parameter_prefix}/admin-password"
  description = "VPSentinel administrator password, published by the control plane"
  type        = "SecureString"
  value       = local.pending_placeholder

  lifecycle {
    ignore_changes = [value]
  }
}

resource "aws_ssm_parameter" "agent_token" {
  count = var.agent_count

  name        = "${local.parameter_prefix}/agent-tokens/${local.node_names[count.index]}"
  description = "Agent token for ${local.node_names[count.index]}, published by the control plane"
  type        = "SecureString"
  value       = local.pending_placeholder

  lifecycle {
    ignore_changes = [value]
  }
}
