variable "service_name" {
  type        = string
  description = "Base name for ALB resources"
}

variable "container_port" {
  type        = number
  description = "Port your app listens on"
}

variable "subnet_ids" {
  type        = list(string)
  description = "Subnets for ALB"
}

variable "security_group_ids" {
  type        = list(string)
  description = "Security groups for ALB"
}

variable "vpc_id" {
  type        = string
  description = "VPC ID for target group"
}