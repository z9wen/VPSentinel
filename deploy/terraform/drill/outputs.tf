output "control_plane_instance_id" {
  description = "Instance running the VPSentinel stack."
  value       = aws_instance.control_plane.id
}

output "agent_instance_ids" {
  description = "Drill node name to instance ID."
  value       = zipmap(local.node_names, aws_instance.agent[*].id)
}

output "console_tunnel_command" {
  description = "Forward the console to http://127.0.0.1:18080 through Session Manager."
  value       = "aws ssm start-session --region ${var.region} --target ${aws_instance.control_plane.id} --document-name AWS-StartPortForwardingSessionToRemoteHost --parameters '{\"host\":[\"${aws_instance.control_plane.private_ip}\"],\"portNumber\":[\"8080\"],\"localPortNumber\":[\"18080\"]}'"
}

output "admin_password_command" {
  description = "Print the administrator password (user admin@example.com)."
  value       = "aws ssm get-parameter --region ${var.region} --name ${aws_ssm_parameter.admin_password.name} --with-decryption --query Parameter.Value --output text"
}

output "bootstrap_log_command" {
  description = "Follow the control plane bootstrap log."
  value       = "aws ssm start-session --region ${var.region} --target ${aws_instance.control_plane.id} --document-name AWS-StartInteractiveCommand --parameters '{\"command\":[\"sudo tail -n 100 -f /var/log/vpsentinel-bootstrap.log\"]}'"
}
