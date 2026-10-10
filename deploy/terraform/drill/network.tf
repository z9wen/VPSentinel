# A single public subnet keeps the drill cheap: instances reach GitHub, GHCR,
# and the SSM endpoints through the internet gateway, so no NAT gateway is
# needed. Nothing accepts inbound traffic from the internet; operators connect
# through SSM Session Manager.

# Pick an availability zone that offers both instance types.
data "aws_ec2_instance_type_offerings" "control_plane" {
  location_type = "availability-zone"

  filter {
    name   = "instance-type"
    values = [var.control_plane_instance_type]
  }
}

data "aws_ec2_instance_type_offerings" "agent" {
  location_type = "availability-zone"

  filter {
    name   = "instance-type"
    values = [var.agent_instance_type]
  }
}

locals {
  availability_zone = sort(setintersection(
    data.aws_ec2_instance_type_offerings.control_plane.locations,
    data.aws_ec2_instance_type_offerings.agent.locations,
  ))[0]
}

resource "aws_vpc" "this" {
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = { Name = var.name }
}

resource "aws_internet_gateway" "this" {
  vpc_id = aws_vpc.this.id

  tags = { Name = var.name }
}

resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.this.id
  cidr_block              = cidrsubnet(var.vpc_cidr, 8, 0)
  availability_zone       = local.availability_zone
  map_public_ip_on_launch = true

  tags = { Name = "${var.name}-public" }
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.this.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.this.id
  }

  tags = { Name = "${var.name}-public" }
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}

resource "aws_security_group" "control_plane" {
  name_prefix = "${var.name}-control-plane-"
  description = "VPSentinel control plane: agent ingestion from drill nodes only"
  vpc_id      = aws_vpc.this.id

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_security_group" "agent" {
  name_prefix = "${var.name}-agent-"
  description = "VPSentinel drill nodes: no inbound access"
  vpc_id      = aws_vpc.this.id

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_vpc_security_group_ingress_rule" "agents_to_control_plane" {
  security_group_id            = aws_security_group.control_plane.id
  referenced_security_group_id = aws_security_group.agent.id
  ip_protocol                  = "tcp"
  from_port                    = 8080
  to_port                      = 8080
  description                  = "Agent metric ingestion over the VPC"
}

# Egress is limited to HTTPS (package repositories, GitHub, GHCR, Docker Hub,
# SSM) plus agent-to-control-plane ingestion.
resource "aws_vpc_security_group_egress_rule" "control_plane_https" {
  security_group_id = aws_security_group.control_plane.id
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "tcp"
  from_port         = 443
  to_port           = 443
  description       = "HTTPS to package repositories, registries, and AWS APIs"
}

resource "aws_vpc_security_group_egress_rule" "agent_https" {
  security_group_id = aws_security_group.agent.id
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "tcp"
  from_port         = 443
  to_port           = 443
  description       = "HTTPS to GitHub releases and AWS APIs"
}

resource "aws_vpc_security_group_egress_rule" "agent_to_control_plane" {
  security_group_id            = aws_security_group.agent.id
  referenced_security_group_id = aws_security_group.control_plane.id
  ip_protocol                  = "tcp"
  from_port                    = 8080
  to_port                      = 8080
  description                  = "Metric ingestion"
}
