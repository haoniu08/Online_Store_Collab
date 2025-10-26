# SQS Queue for order processing
# Configuration as specified in assignment:
# - Visibility timeout: 30 seconds (default)
# - Message retention: 4 days (default)
# - Receive wait time: 20 seconds (long polling)

resource "aws_sqs_queue" "order_processing" {
  name = "${var.service_name}-order-processing-queue"

  # Assignment specifications
  visibility_timeout_seconds = 30  # Default, time for worker to process before message becomes visible again
  message_retention_seconds  = 345600  # 4 days (96 hours)
  receive_wait_time_seconds  = 20  # Long polling - wait up to 20s for messages

  # Additional settings for production readiness
  delay_seconds              = 0   # No delay
  max_message_size          = 262144  # 256 KB (default)

  tags = {
    Name        = "${var.service_name}-order-processing-queue"
    Environment = var.environment
    Purpose     = "Order processing queue"
  }
}

# Subscribe SQS to SNS topic
resource "aws_sns_topic_subscription" "order_processing_subscription" {
  topic_arn = var.sns_topic_arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.order_processing.arn

  # No filtering - receive all messages from SNS
  raw_message_delivery = true  # Deliver SNS message as-is without SNS envelope
}

# SQS Queue Policy to allow SNS to send messages
resource "aws_sqs_queue_policy" "order_processing_policy" {
  queue_url = aws_sqs_queue.order_processing.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowSNSToSendMessages"
        Effect = "Allow"
        Principal = {
          Service = "sns.amazonaws.com"
        }
        Action   = "SQS:SendMessage"
        Resource = aws_sqs_queue.order_processing.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = var.sns_topic_arn
          }
        }
      }
    ]
  })
}
