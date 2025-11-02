output "ecs_cluster_name" {
  description = "Name of the created ECS cluster"
  value       = module.ecs.cluster_name
}

output "ecs_service_name" {
  description = "Name of the running ECS service"
  value       = module.ecs.service_name
}

output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = module.alb.alb_dns_name
}

output "load_balancer_url" {
  description = "Complete URL to access your service"
  value       = "http://${module.alb.alb_dns_name}"
}

# RDS MySQL connection details (Homework 8)
output "rds_endpoint" {
  description = "RDS MySQL endpoint"
  value       = module.rds.endpoint
}

output "rds_port" {
  description = "RDS MySQL port"
  value       = module.rds.port
}

output "rds_database_name" {
  description = "RDS database name"
  value       = module.rds.db_name
}

output "rds_username" {
  description = "RDS username"
  value       = module.rds.db_username
}

output "rds_password" {
  description = "RDS password (sensitive)"
  value       = module.rds.db_password
  sensitive   = true
}

output "rds_security_group_id" {
  description = "RDS security group ID"
  value       = module.rds.security_group_id
}