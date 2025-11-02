output "endpoint" {
  description = "RDS endpoint hostname"
  value       = aws_db_instance.this.address
}

output "port" {
  description = "RDS port"
  value       = aws_db_instance.this.port
}

output "security_group_id" {
  description = "RDS security group ID"
  value       = aws_security_group.rds.id
}

output "db_name" {
  description = "Database name"
  value       = var.db_name
}

output "db_username" {
  description = "Database master username"
  value       = var.db_username
}

output "db_password" {
  description = "Generated database password (sensitive)"
  value       = random_password.db_password.result
  sensitive   = true
}


