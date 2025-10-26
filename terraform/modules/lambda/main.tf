// Lambda module: deploy image-based Lambda, SNS subscription, permission, and DLQ
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 3.0"
    }
  }
}

resource "aws_sqs_queue" "lambda_dlq" {
  name                      = var.dlq_name
  visibility_timeout_seconds = var.dlq_visibility_timeout
  message_retention_seconds  = var.dlq_retention_seconds
  tags = var.tags
}

resource "aws_lambda_function" "this" {
  function_name = var.function_name
  package_type  = "Image"
  image_uri     = var.image_uri
  role          = var.lambda_role_arn
  memory_size   = var.memory_size
  timeout       = var.timeout
  dead_letter_config {
    target_arn = aws_sqs_queue.lambda_dlq.arn
  }

  tags = var.tags
}

resource "aws_lambda_permission" "allow_sns" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.this.arn
  principal     = "sns.amazonaws.com"
  source_arn    = var.sns_topic_arn
}

resource "aws_sns_topic_subscription" "lambda_sub" {
  topic_arn = var.sns_topic_arn
  protocol  = "lambda"
  endpoint  = aws_lambda_function.this.arn
}

// Note: dead-letter queue is configured on the function via dead_letter_config.
