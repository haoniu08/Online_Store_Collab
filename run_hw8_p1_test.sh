#!/bin/bash
# HW8 Testing Script - Step by step execution
# This script rebuilds/deploys, applies migrations, and runs Locust tests

set -e  # Exit on error

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TERRAFORM_DIR="$SCRIPT_DIR/terraform"
PROJECT_ROOT="$SCRIPT_DIR"

echo "=========================================="
echo "HW8 MySQL Shopping Cart Testing Script"
echo "=========================================="
echo ""

# Step 1: Navigate to terraform directory and get outputs
echo "Step 1: Getting Terraform outputs..."
cd "$TERRAFORM_DIR"

# Set AWS region
export AWS_REGION="us-west-2"

# Get all RDS connection details (strip newlines)
RDS_ENDPOINT=$(terraform output -raw rds_endpoint | tr -d '\n\r')
RDS_USER=$(terraform output -raw rds_username | tr -d '\n\r')
RDS_DB=$(terraform output -raw rds_database_name | tr -d '\n\r')
RDS_PWD=$(terraform output -raw rds_password | tr -d '\n\r')
RDS_SG_ID=$(terraform output -raw rds_security_group_id | tr -d '\n\r')
BASE_URL=$(terraform output -raw load_balancer_url | tr -d '\n\r')

echo "  ✓ RDS Endpoint: $RDS_ENDPOINT"
echo "  ✓ RDS User: $RDS_USER"
echo "  ✓ RDS Database: $RDS_DB"
echo "  ✓ RDS SG ID: $RDS_SG_ID"
echo "  ✓ Base URL: $BASE_URL"
echo ""

# Step 2: Rebuild and redeploy ECS service
echo "Step 2: Rebuilding and redeploying ECS service..."
echo "  - Tainting Docker resources to force rebuild..."

terraform taint -allow-missing docker_image.app 2>/dev/null || true
terraform taint -allow-missing docker_registry_image.app 2>/dev/null || true

echo "  - Applying Terraform (this will rebuild Docker image)..."
terraform apply -auto-approve

echo "  - Forcing ECS service deployment to pick up new image..."
CLUSTER_NAME=$(terraform output -raw ecs_cluster_name | tr -d '\n\r')
SERVICE_NAME=$(terraform output -raw ecs_service_name | tr -d '\n\r')

aws ecs update-service \
  --cluster "$CLUSTER_NAME" \
  --service "$SERVICE_NAME" \
  --force-new-deployment \
  --region "$AWS_REGION" > /dev/null

echo "  ✓ Deployment initiated. Waiting 60 seconds for service to stabilize..."
sleep 60
echo ""

# Step 3: Verify migrations (they run automatically on app startup)
echo "Step 3: Verifying database setup..."
echo "  ✓ Migrations run automatically on app startup (embedded in main.go)"
echo "  - Waiting for app to initialize and run migrations..."
echo "  - You can check ECS logs to confirm: aws logs tail /ecs/CS6650L2 --follow"
echo ""

# Step 4: Verify service is responding
echo "Step 4: Verifying service health..."
HEALTH_CHECK=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health" || echo "000")
if [ "$HEALTH_CHECK" = "200" ]; then
  echo "  ✓ Service is healthy (HTTP $HEALTH_CHECK)"
else
  echo "  ⚠ Service health check returned: $HEALTH_CHECK"
  echo "  Waiting additional 30 seconds for service to fully start..."
  sleep 30
fi
echo ""

# Step 5: Quick smoke test
echo "Step 5: Running smoke test..."
SMOKE_RESPONSE=$(curl -s -X POST "$BASE_URL/shopping-carts" \
  -H 'Content-Type: application/json' \
  -d '{"customer_id":1}' 2>&1)

if echo "$SMOKE_RESPONSE" | grep -q "shopping_cart_id"; then
  CART_ID=$(echo "$SMOKE_RESPONSE" | grep -o '"shopping_cart_id":[0-9]*' | cut -d: -f2)
  echo "  ✓ Created test cart: ID $CART_ID"
  
  # Test add item
  ADD_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/shopping-carts/$CART_ID/items" \
    -H 'Content-Type: application/json' \
    -d '{"product_id":123,"quantity":2}')
  if [ "$ADD_RESPONSE" = "204" ]; then
    echo "  ✓ Added item to cart (HTTP $ADD_RESPONSE)"
  fi
  
  # Test get cart
  GET_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/shopping-carts/$CART_ID")
  if [ "$GET_RESPONSE" = "200" ]; then
    echo "  ✓ Retrieved cart (HTTP $GET_RESPONSE)"
  fi
else
  echo "  ⚠ Smoke test failed. Response: $SMOKE_RESPONSE"
  echo "  Proceeding anyway - service may still be starting..."
fi
echo ""

# Step 6: Run Locust tests
echo "Step 6: Running Locust performance tests..."
cd "$PROJECT_ROOT/test_locust_hw8"

export LOCUST_HOST="$BASE_URL"
echo "  - Target host: $LOCUST_HOST"
echo "  - Running 150 operations (50 create, 50 add, 50 get)..."
echo ""

# Check if locust is installed
if ! command -v locust &> /dev/null; then
  echo "  ⚠ Locust not found. Using Docker Compose..."
  cd "$PROJECT_ROOT/test_locust_hw8"
  LOCUST_HOST="$BASE_URL" docker compose run --rm locust
  cd "$PROJECT_ROOT"
else
  locust -f locustfile.py \
    --host "$LOCUST_HOST" \
    --headless \
    -u 1 \
    -r 1 \
    --run-time 10m \
    --stop-timeout 30 \
    --html report.html \
    --csv results
fi

# Step 7: Move results to reports directory
echo ""
echo "Step 7: Saving test results..."
if [ -f "mysql_test_results.json" ]; then
  mkdir -p "$PROJECT_ROOT/reports/HW8"
  mv mysql_test_results.json "$PROJECT_ROOT/reports/HW8/mysql_test_results.json"
  echo "  ✓ Results saved to reports/HW8/mysql_test_results.json"
  
  # Count operations
  CREATE_COUNT=$(grep -c '"operation":"create_cart"' "$PROJECT_ROOT/reports/HW8/mysql_test_results.json" || echo "0")
  ADD_COUNT=$(grep -c '"operation":"add_items"' "$PROJECT_ROOT/reports/HW8/mysql_test_results.json" || echo "0")
  GET_COUNT=$(grep -c '"operation":"get_cart"' "$PROJECT_ROOT/reports/HW8/mysql_test_results.json" || echo "0")
  
  echo "  - Operations completed:"
  echo "    Creates: $CREATE_COUNT/50"
  echo "    Adds: $ADD_COUNT/50"
  echo "    Gets: $GET_COUNT/50"
else
  echo "  ⚠ mysql_test_results.json not found. Check Locust output above."
fi

# Step 8: Run analysis if script exists
if [ -f "analyze_results.py" ] && [ -f "$PROJECT_ROOT/reports/HW8/mysql_test_results.json" ]; then
  echo ""
  echo "Step 8: Analyzing results..."
  python3 analyze_results.py "$PROJECT_ROOT/reports/HW8/mysql_test_results.json" || true
fi

echo ""
echo "=========================================="
echo "Testing complete!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "  1. Check reports/HW8/mysql_test_results.json"
echo "  2. Review CloudWatch metrics for RDS and ECS"
echo "  3. Check ECS logs: aws logs tail /ecs/CS6650L2 --follow"
echo ""
