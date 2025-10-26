output "service_name" {
  description = "Name of the processor ECS service"
  value       = aws_ecs_service.processor.name
}

output "task_definition_arn" {
  description = "ARN of the processor task definition"
  value       = aws_ecs_task_definition.processor.arn
}
