#!/bin/bash
# Script to query RDS MySQL database
# NOTE: This script will NOT work if RDS is in private subnets (default security setup)
# RDS security group only allows connections from ECS tasks, not external IPs
# 
# For querying RDS in private subnets, see:
# - query_rds_via_ecs.sh (runs query via ECS task)
# - Or verify via API endpoints and ECS logs
#
# Temporarily opens security group, runs queries, then closes it

set -e

cd "$(dirname "$0")/terraform"

# Get connection details
RDS_ENDPOINT=$(terraform output -raw rds_endpoint | tr -d '\n\r')
RDS_USER=$(terraform output -raw rds_username | tr -d '\n\r')
RDS_DB=$(terraform output -raw rds_database_name | tr -d '\n\r')
RDS_PWD=$(terraform output -raw rds_password | tr -d '\n\r')
RDS_SG_ID=$(terraform output -raw rds_security_group_id | tr -d '\n\r')
MY_IP=$(curl -s https://checkip.amazonaws.com | tr -d '\n\r')

echo "=========================================="
echo "RDS MySQL Query Tool"
echo "=========================================="
echo "⚠ WARNING: This script assumes RDS allows public access or your IP is allowed."
echo "If RDS is in private subnets (default), this will fail."
echo "Use query_rds_via_ecs.sh instead, or verify via API/ECS logs."
echo ""
echo "Endpoint: $RDS_ENDPOINT"
echo "Database: $RDS_DB"
echo "User: $RDS_USER"
echo "Security Group: $RDS_SG_ID"
echo ""

# Open security group temporarily
echo "Opening RDS security group for your IP ($MY_IP)..."
aws ec2 authorize-security-group-ingress \
  --group-id "$RDS_SG_ID" \
  --protocol tcp \
  --port 3306 \
  --cidr "$MY_IP/32" \
  --region us-west-2 2>/dev/null || echo "  (Rule may already exist)"

echo "✓ Security group opened"
echo ""

# Check if mysql client is available
if ! command -v mysql &> /dev/null; then
  echo "⚠ mysql client not found. Install with:"
  echo "   macOS: brew install mysql-client"
  echo "   Ubuntu: sudo apt-get install mysql-client"
  echo ""
  echo "Connection string (use with your MySQL client):"
  echo "  mysql -h $RDS_ENDPOINT -P 3306 -u $RDS_USER -p$RDS_PWD $RDS_DB"
  echo ""
  read -p "Press Enter to close security group..."
else
  echo "Running common queries..."
  echo ""
  
  # Run useful queries
  mysql -h "$RDS_ENDPOINT" -P 3306 -u "$RDS_USER" -p"$RDS_PWD" "$RDS_DB" << 'EOF'
-- Show all tables
SELECT "=== TABLES ===" AS info;
SHOW TABLES;

-- Describe shopping_carts table
SELECT "\n=== shopping_carts SCHEMA ===" AS info;
DESCRIBE shopping_carts;

-- Describe shopping_cart_items table
SELECT "\n=== shopping_cart_items SCHEMA ===" AS info;
DESCRIBE shopping_cart_items;

-- Show indexes
SELECT "\n=== INDEXES ===" AS info;
SHOW INDEX FROM shopping_carts;
SHOW INDEX FROM shopping_cart_items;

-- Count rows
SELECT "\n=== ROW COUNTS ===" AS info;
SELECT 
  (SELECT COUNT(*) FROM shopping_carts) AS total_carts,
  (SELECT COUNT(*) FROM shopping_cart_items) AS total_items;

-- Sample data (first 5 carts)
SELECT "\n=== SAMPLE CARTS (first 5) ===" AS info;
SELECT * FROM shopping_carts LIMIT 5;

-- Sample items (first 10)
SELECT "\n=== SAMPLE ITEMS (first 10) ===" AS info;
SELECT * FROM shopping_cart_items LIMIT 10;

-- Cart summary with item counts
SELECT "\n=== CART SUMMARY (with item counts) ===" AS info;
SELECT 
  c.shopping_cart_id,
  c.customer_id,
  c.status,
  COUNT(i.product_id) AS item_count,
  SUM(i.quantity) AS total_quantity,
  c.created_at
FROM shopping_carts c
LEFT JOIN shopping_cart_items i ON c.shopping_cart_id = i.shopping_cart_id
GROUP BY c.shopping_cart_id, c.customer_id, c.status, c.created_at
ORDER BY c.shopping_cart_id DESC
LIMIT 10;

EOF

  echo ""
  echo "=========================================="
  echo "Queries completed!"
  echo "=========================================="
fi

# Close security group
echo ""
echo "Closing RDS security group..."
aws ec2 revoke-security-group-ingress \
  --group-id "$RDS_SG_ID" \
  --protocol tcp \
  --port 3306 \
  --cidr "$MY_IP/32" \
  --region us-west-2 2>/dev/null || true

echo "✓ Security group closed"
echo ""
echo "To run custom queries, you can:"
echo "  1. Run this script again and edit the MySQL queries section"
echo "  2. Or connect manually:"
echo "     mysql -h $RDS_ENDPOINT -P 3306 -u $RDS_USER -p'$RDS_PWD' $RDS_DB"

