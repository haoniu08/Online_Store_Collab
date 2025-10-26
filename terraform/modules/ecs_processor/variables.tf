variable "service_name" {
  type        = string
  description = "Base name for ECS resources"
}

variable "cluster_id" {
  type        = string
  description = "ID of the ECS cluster"
}

variable "image" {
  type        = string
  description = "ECR image URI (with tag)"
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

variable "desired_count" {
  type        = number
  default     = 1
  description = "Desired number of processor tasks"
}

variable "region" {
  type        = string
  description = "AWS region"
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

variable "sqs_queue_url" {
  type        = string
  description = "URL of SQS queue for order processing"
}

variable "worker_count" {
  type        = number
  default     = 1
  description = "Number of concurrent worker goroutines (for Phase 5 scaling)"
}
