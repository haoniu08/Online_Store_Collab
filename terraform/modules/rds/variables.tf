variable "service_name" {
  type        = string
  description = "Base name for RDS resources"
}

variable "vpc_id" {
  type        = string
  description = "VPC ID for RDS security group"
}

variable "subnet_ids" {
  type        = list(string)
  description = "Subnet IDs for DB subnet group (private preferred)"
}

variable "allowed_security_group_ids" {
  type        = list(string)
  description = "Security group IDs allowed to access MySQL (e.g., ECS tasks)"
}

variable "db_name" {
  type        = string
  description = "Database name"
  default     = "appdb"
}

variable "db_username" {
  type        = string
  description = "Master username"
  default     = "appuser"
}


