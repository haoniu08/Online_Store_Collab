#!/bin/bash
# HW8 Step 2 - DynamoDB Performance Testing Script
# Runs 150 operations against DynamoDB-backed API

set -e  # Exit on error

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TERRAFORM_DIR="$SCRIPT_DIR/terraform"
PROJECT_ROOT="$SCRIPT_DIR"

echo "=========================================="
echo "HW8 DynamoDB Shopping Cart Testing Script"
echo "=========================================="
echo ""

# Step 1: Get Terraform outputs
echo "Step 1: Getting infrastructure details..."
cd "$TERRAFORM_DIR"

BASE_URL=$(terraform output -raw load_balancer_url | tr -d '\n\r')
echo "  ✓ Base URL: $BASE_URL"
echo ""

# Step 2: Verify DynamoDB is being used
echo "Step 2: Verifying DynamoDB configuration..."
DYNAMODB_TABLE=$(terraform output -raw dynamodb_table_name 2>/dev/null | tr -d '\n\r' || echo "N/A")
if [ "$DYNAMODB_TABLE" != "N/A" ]; then
  echo "  ✓ DynamoDB Table: $DYNAMODB_TABLE"

  # Check table status
  TABLE_STATUS=$(aws dynamodb describe-table --table-name "$DYNAMODB_TABLE" --query "Table.TableStatus" --output text 2>/dev/null || echo "UNKNOWN")
  echo "  ✓ Table Status: $TABLE_STATUS"
else
  echo "  ⚠ Warning: Could not find DynamoDB table name in terraform outputs"
fi
echo ""

# Step 3: Verify service health
echo "Step 3: Verifying service health..."
HEALTH_CHECK=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health" || echo "000")
if [ "$HEALTH_CHECK" = "200" ]; then
  echo "  ✓ Service is healthy (HTTP $HEALTH_CHECK)"
else
  echo "  ⚠ Service health check returned: $HEALTH_CHECK"
  echo "  Waiting 30 seconds for service to stabilize..."
  sleep 30
fi
echo ""

# Step 4: Quick smoke test
echo "Step 4: Running smoke test with DynamoDB..."
SMOKE_RESPONSE=$(curl -s -X POST "$BASE_URL/shopping-carts" \
  -H 'Content-Type: application/json' \
  -d '{"customer_id":99999}' 2>&1)

if echo "$SMOKE_RESPONSE" | grep -q "shopping_cart_id"; then
  CART_ID=$(echo "$SMOKE_RESPONSE" | grep -o '"shopping_cart_id":[0-9]*' | cut -d: -f2)
  echo "  ✓ Created test cart: ID $CART_ID"

  # Verify it's in DynamoDB
  if [ "$DYNAMODB_TABLE" != "N/A" ]; then
    ITEM_EXISTS=$(aws dynamodb get-item --table-name "$DYNAMODB_TABLE" --key "{\"shopping_cart_id\":{\"N\":\"$CART_ID\"}}" --query "Item.shopping_cart_id.N" --output text 2>/dev/null || echo "NOT_FOUND")
    if [ "$ITEM_EXISTS" = "$CART_ID" ]; then
      echo "  ✓ Verified cart exists in DynamoDB"
    fi
  fi

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
  echo "  Service may not be using DynamoDB or may still be starting..."
fi
echo ""

# Step 5: Run Locust tests
echo "Step 5: Running Locust performance tests..."
cd "$PROJECT_ROOT/test_locust_hw8"

export LOCUST_HOST="$BASE_URL"
echo "  - Target host: $LOCUST_HOST"
echo "  - Running 150 operations (50 create, 50 add, 50 get)..."
echo ""

# Check if locust is installed
if ! command -v locust &> /dev/null; then
  echo "  ⚠ Locust not found. Installing..."
  echo "  Please install locust: pip install locust"
  exit 1
fi

locust -f locustfile_dynamodb.py \
  --host "$LOCUST_HOST" \
  --headless \
  -u 1 \
  -r 1 \
  --run-time 5m \
  --stop-timeout 30

# Step 6: Move results to reports directory
echo ""
echo "Step 6: Saving test results..."
if [ -f "dynamodb_test_results.json" ]; then
  mkdir -p "$PROJECT_ROOT/reports/HW8"
  mv dynamodb_test_results.json "$PROJECT_ROOT/reports/HW8/dynamodb_test_results.json"
  echo "  ✓ Results saved to reports/HW8/dynamodb_test_results.json"

  # Count operations
  CREATE_COUNT=$(grep -c '"operation":"create_cart"' "$PROJECT_ROOT/reports/HW8/dynamodb_test_results.json" || echo "0")
  ADD_COUNT=$(grep -c '"operation":"add_items"' "$PROJECT_ROOT/reports/HW8/dynamodb_test_results.json" || echo "0")
  GET_COUNT=$(grep -c '"operation":"get_cart"' "$PROJECT_ROOT/reports/HW8/dynamodb_test_results.json" || echo "0")

  echo "  - Operations completed:"
  echo "    Creates: $CREATE_COUNT/50"
  echo "    Adds: $ADD_COUNT/50"
  echo "    Gets: $GET_COUNT/50"

  # Quick performance summary
  AVG_RESPONSE=$(jq '[.[] | .response_time] | add / length' "$PROJECT_ROOT/reports/HW8/dynamodb_test_results.json" 2>/dev/null || echo "N/A")
  if [ "$AVG_RESPONSE" != "N/A" ]; then
    echo "  - Average response time: ${AVG_RESPONSE}ms"
  fi
else
  echo "  ⚠ dynamodb_test_results.json not found. Check Locust output above."
fi

echo ""
echo "=========================================="
echo "Testing complete!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "  1. Check reports/HW8/dynamodb_test_results.json"
echo "  2. Compare with reports/HW8/mysql_test_results.json"
echo "  3. Review CloudWatch metrics for DynamoDB"
echo "  4. Proceed to Step 3 (comparison analysis)"
echo ""
