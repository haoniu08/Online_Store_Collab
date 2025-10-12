variable "service_name" {
  type        = string
  description = "Base name for ECS resources"
}

variable "image" {
  type        = string
  description = "ECR image URI (with tag)"
}

variable "container_port" {
  type        = number
  description = "Port your app listens on"
}

variable "subnet_ids" {
  type        = list(string)
  description = "Subnets for FARGATE tasks"
}

variable "security_group_ids" {
  type        = list(string)
  description = "SGs for FARGATE tasks"
}

variable "execution_role_arn" {
  type        = string
  description = "ECS Task Execution Role ARN"
}

variable "task_role_arn" {
  type        = string
  description = "IAM Role ARN for app permissions"
}

variable "log_group_name" {
  type        = string
  description = "CloudWatch log group name"
}

variable "ecs_count" {
  type        = number
  default     = 1
  description = "Desired Fargate task count"
}

variable "region" {
  type        = string
  description = "AWS region (for awslogs driver)"
}

variable "cpu" {
  type        = string
  default     = "256"
  description = "vCPU units"
}

variable "memory" {
  type        = string
  default     = "512"
  description = "Memory (MiB)"
}

variable "target_group_arn" {
  type        = string
  description = "ARN of the ALB target group"
}

variable "alb_listener_arn" {
  type        = string
  description = "ARN of the ALB listener (for dependency)"
}

variable "min_capacity" {
  type        = number
  default     = 2
  description = "Minimum number of ECS tasks"
}

variable "max_capacity" {
  type        = number
  default     = 4
  description = "Maximum number of ECS tasks"
}

variable "cpu_target_percentage" {
  type        = number
  default     = 70
  description = "Target CPU utilization percentage for auto scaling"
}

variable "scale_out_cooldown" {
  type        = number
  default     = 300
  description = "Scale out cooldown in seconds"
}

variable "scale_in_cooldown" {
  type        = number
  default     = 300
  description = "Scale in cooldown in seconds"
}
