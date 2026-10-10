data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}

locals {
  parameter_arn_prefix = "arn:${data.aws_partition.current.partition}:ssm:${var.region}:${data.aws_caller_identity.current.account_id}:parameter${local.parameter_prefix}"
  ssm_core_policy_arn  = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

data "aws_iam_policy_document" "ec2_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

# Control plane: Session Manager access plus permission to publish the admin
# password and the per-node agent tokens it creates during bootstrap.
resource "aws_iam_role" "control_plane" {
  name_prefix        = "${var.name}-cp-"
  assume_role_policy = data.aws_iam_policy_document.ec2_assume_role.json
}

resource "aws_iam_role_policy_attachment" "control_plane_ssm" {
  role       = aws_iam_role.control_plane.name
  policy_arn = local.ssm_core_policy_arn
}

data "aws_iam_policy_document" "control_plane_parameters" {
  statement {
    actions   = ["ssm:PutParameter"]
    resources = ["${local.parameter_arn_prefix}/*"]
  }
}

resource "aws_iam_role_policy" "control_plane_parameters" {
  name   = "publish-bootstrap-secrets"
  role   = aws_iam_role.control_plane.id
  policy = data.aws_iam_policy_document.control_plane_parameters.json
}

resource "aws_iam_instance_profile" "control_plane" {
  name_prefix = "${var.name}-cp-"
  role        = aws_iam_role.control_plane.name
}

# One role per node so each instance can read only its own agent token.
resource "aws_iam_role" "agent" {
  count = var.agent_count

  name_prefix        = "${var.name}-agent${count.index + 1}-"
  assume_role_policy = data.aws_iam_policy_document.ec2_assume_role.json
}

resource "aws_iam_role_policy_attachment" "agent_ssm" {
  count = var.agent_count

  role       = aws_iam_role.agent[count.index].name
  policy_arn = local.ssm_core_policy_arn
}

data "aws_iam_policy_document" "agent_token" {
  count = var.agent_count

  statement {
    actions   = ["ssm:GetParameter"]
    resources = [aws_ssm_parameter.agent_token[count.index].arn]
  }
}

resource "aws_iam_role_policy" "agent_token" {
  count = var.agent_count

  name   = "read-own-agent-token"
  role   = aws_iam_role.agent[count.index].id
  policy = data.aws_iam_policy_document.agent_token[count.index].json
}

resource "aws_iam_instance_profile" "agent" {
  count = var.agent_count

  name_prefix = "${var.name}-agent${count.index + 1}-"
  role        = aws_iam_role.agent[count.index].name
}
