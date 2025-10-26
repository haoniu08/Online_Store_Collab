# SNS Topic for order processing events
# This topic receives order events from the API and distributes them to subscribers (SQS)

resource "aws_sns_topic" "order_processing" {
  name = "${var.service_name}-order-processing-events"

  tags = {
    Name        = "${var.service_name}-order-processing"
    Environment = var.environment
    Purpose     = "Order processing event distribution"
  }
}

# SNS Topic Policy to allow publishing from ECS tasks
resource "aws_sns_topic_policy" "order_processing_policy" {
  arn = aws_sns_topic.order_processing.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowECSTasksToPublish"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action = [
          "SNS:Publish"
        ]
        Resource = aws_sns_topic.order_processing.arn
      }
    ]
  })
}
