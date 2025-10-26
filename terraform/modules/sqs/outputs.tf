output "queue_url" {
  description = "URL of the SQS queue"
  value       = aws_sqs_queue.order_processing.url
}

output "queue_arn" {
  description = "ARN of the SQS queue"
  value       = aws_sqs_queue.order_processing.arn
}

output "queue_name" {
  description = "Name of the SQS queue"
  value       = aws_sqs_queue.order_processing.name
}
