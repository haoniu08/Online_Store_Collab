# ECS Service for Order Processor (background worker)
# This service polls SQS and processes orders

# ECS Cluster - reuse the same cluster
# (passed as variable from main.tf)

# Task Definition for Order Processor
resource "aws_ecs_task_definition" "processor" {
  family                   = "${var.service_name}-processor-task"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.cpu
  memory                   = var.memory

  execution_role_arn = var.execution_role_arn
  task_role_arn      = var.task_role_arn

  container_definitions = jsonencode([{
    name      = "${var.service_name}-processor-container"
    image     = var.image
    essential = true

    environment = [
      {
        name  = "AWS_REGION"
        value = var.region
      },
      {
        name  = "SQS_QUEUE_URL"
        value = var.sqs_queue_url
      },
      {
        name  = "WORKER_COUNT"
        value = tostring(var.worker_count)
      }
    ]

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = var.log_group_name
        "awslogs-region"        = var.region
        "awslogs-stream-prefix" = "processor"
      }
    }
  }])
}

# ECS Service for Processor (no load balancer needed)
resource "aws_ecs_service" "processor" {
  name            = "${var.service_name}-processor"
  cluster         = var.cluster_id
  task_definition = aws_ecs_task_definition.processor.arn
  desired_count   = var.desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.subnet_ids
    security_groups  = var.security_group_ids
    assign_public_ip = true
  }

  # Processor doesn't need load balancer - it's a background worker
}
