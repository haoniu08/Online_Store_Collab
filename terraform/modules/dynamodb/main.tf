# DynamoDB table for shopping carts
# Uses single-table design with embedded items for optimal performance

resource "aws_dynamodb_table" "shopping_carts" {
  name         = "${var.project_name}-shopping-carts-${var.environment}"
  billing_mode = "PAY_PER_REQUEST" # On-demand pricing, no capacity planning needed

  # Partition key: shopping_cart_id - ensures even distribution
  hash_key = "shopping_cart_id"

  attribute {
    name = "shopping_cart_id"
    type = "N" # Number type
  }

  # Optional: Global Secondary Index for customer queries
  # Commented out for now as assignment doesn't require it
  # attribute {
  #   name = "customer_id"
  #   type = "N"
  # }
  #
  # global_secondary_index {
  #   name            = "customer-index"
  #   hash_key        = "customer_id"
  #   projection_type = "ALL"
  # }

  # Enable point-in-time recovery for production (optional for assignment)
  point_in_time_recovery {
    enabled = var.enable_point_in_time_recovery
  }

  # Server-side encryption
  server_side_encryption {
    enabled = true
  }

  tags = {
    Name        = "${var.project_name}-shopping-carts-${var.environment}"
    Environment = var.environment
    ManagedBy   = "Terraform"
    Assignment  = "HW8-Step2"
  }
}
