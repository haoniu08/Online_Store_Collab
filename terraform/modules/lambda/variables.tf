variable "image_uri" {
  description = "ECR image URI (including tag) for the Lambda image"
  type        = string
}

variable "function_name" {
  description = "Name for the Lambda function"
  type        = string
}

variable "lambda_role_arn" {
  description = "IAM role ARN for Lambda execution (must allow logs/write and any AWS calls needed)"
  type        = string
}

variable "sns_topic_arn" {
  description = "SNS topic ARN to subscribe the Lambda to"
  type        = string
}

variable "memory_size" {
  description = "Lambda memory size in MB"
  type        = number
  default     = 512
}

variable "timeout" {
  description = "Lambda timeout in seconds"
  type        = number
  default     = 30
}

variable "dlq_name" {
  description = "SQS queue name to use as DLQ for failed Lambda invocations"
  type        = string
  default     = "lambda-dlq"
}

variable "dlq_visibility_timeout" {
  description = "Visibility timeout for the DLQ"
  type        = number
  default     = 60
}

variable "dlq_retention_seconds" {
  description = "Message retention for the DLQ"
  type        = number
  default     = 1209600 # 14 days
}

variable "maximum_retry_attempts" {
  description = "Maximum retry attempts for asynchronous Lambda invocations"
  type        = number
  default     = 2
}

variable "tags" {
  description = "Optional tags applied to resources"
  type        = map(string)
  default     = {}
}
