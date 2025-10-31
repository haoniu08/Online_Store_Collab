# Region to deploy into
variable "aws_region" {
  type    = string
  default = "us-west-2"
}

# ECR & ECS settings
variable "ecr_repository_name" {
  type    = string
  default = "ecr_service"
}

variable "service_name" {
  type    = string
  default = "CS6650L2"
}

variable "container_port" {
  type    = number
  default = 8080
}

variable "ecs_count" {
  type    = number
  default = 1
}

# How long to keep logs
variable "log_retention_days" {
  type    = number
  default = 7
}

# ECS Task CPU and Memory
variable "cpu" {
  type        = string
  default     = "256"
  description = "CPU units for ECS task (256, 512, 1024, 2048, 4096)"
}

variable "memory" {
  type        = string
  default     = "512"
  description = "Memory (MiB) for ECS task"
}

# Auto Scaling Configuration
variable "min_instances" {
  type        = number
  default     = 2
  description = "Minimum number of ECS tasks"
}

variable "max_instances" {
  type        = number
  default     = 4
  description = "Maximum number of ECS tasks"
}

variable "cpu_target_percentage" {
  type        = number
  default     = 70
  description = "Target CPU utilization percentage for auto scaling"
}

# Database settings (Homework 8)
variable "db_name" {
  type        = string
  default     = "appdb"
  description = "MySQL database name"
}

variable "db_username" {
  type        = string
  default     = "appuser"
  description = "MySQL master username"
}
