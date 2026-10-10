data "aws_ssm_parameter" "al2023_arm64" {
  name = "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-arm64"
}

locals {
  # Pinned with its published SHA-256 so bootstrap fails closed on a mismatch.
  compose_version = "v5.5.1"
  compose_sha256  = "732e3a84c1a0f67256ce80bc2598a24546b10ca05f9faa97efceb1171ece2ef7"
}

resource "aws_instance" "control_plane" {
  ami                    = data.aws_ssm_parameter.al2023_arm64.insecure_value
  instance_type          = var.control_plane_instance_type
  subnet_id              = aws_subnet.public.id
  vpc_security_group_ids = [aws_security_group.control_plane.id]
  iam_instance_profile   = aws_iam_instance_profile.control_plane.name

  user_data = templatefile("${path.module}/templates/control-plane.sh.tftpl", {
    region           = var.region
    repository       = var.repository
    release_tag      = var.release_tag
    parameter_prefix = local.parameter_prefix
    node_names       = local.node_names
    compose_version  = local.compose_version
    compose_sha256   = local.compose_sha256
  })
  user_data_replace_on_change = true

  # IMDSv2 only, with a hop limit of 1 so containers cannot reach the
  # instance role credentials through the metadata service.
  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
  }

  root_block_device {
    volume_type = "gp3"
    volume_size = 16
    encrypted   = true
  }

  tags = { Name = "${var.name}-control-plane" }

  # The bootstrap publishes secrets into these parameters, and the role must
  # be usable before the instance boots.
  depends_on = [
    aws_ssm_parameter.admin_password,
    aws_ssm_parameter.agent_token,
    aws_iam_role_policy.control_plane_parameters,
    aws_route_table_association.public,
  ]

  lifecycle {
    ignore_changes = [ami]
  }
}

resource "aws_instance" "agent" {
  count = var.agent_count

  ami                    = data.aws_ssm_parameter.al2023_arm64.insecure_value
  instance_type          = var.agent_instance_type
  subnet_id              = aws_subnet.public.id
  vpc_security_group_ids = [aws_security_group.agent.id]
  iam_instance_profile   = aws_iam_instance_profile.agent[count.index].name

  user_data = templatefile("${path.module}/templates/agent.sh.tftpl", {
    region           = var.region
    repository       = var.repository
    release_tag      = var.release_tag
    node_name        = local.node_names[count.index]
    token_parameter  = aws_ssm_parameter.agent_token[count.index].name
    control_plane_ip = aws_instance.control_plane.private_ip
  })
  user_data_replace_on_change = true

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
  }

  root_block_device {
    volume_type = "gp3"
    encrypted   = true
  }

  tags = { Name = "${var.name}-${local.node_names[count.index]}" }

  depends_on = [aws_iam_role_policy.agent_token]

  lifecycle {
    ignore_changes = [ami]
  }
}
