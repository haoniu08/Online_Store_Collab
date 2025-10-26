output "lambda_function_arn" {
  description = "ARN of the created Lambda function"
  value       = aws_lambda_function.this.arn
}

output "lambda_function_name" {
  description = "Name of the created Lambda function"
  value       = aws_lambda_function.this.function_name
}

output "dlq_url" {
  description = "URL of the DLQ SQS queue"
  value       = aws_sqs_queue.lambda_dlq.id
}

output "dlq_arn" {
  description = "ARN of the DLQ SQS queue"
  value       = aws_sqs_queue.lambda_dlq.arn
}
