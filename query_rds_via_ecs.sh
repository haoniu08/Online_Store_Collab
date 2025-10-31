#!/bin/bash
# Query RDS via ECS task (since RDS is in private subnet)
# This runs a one-off ECS task that connects to RDS and runs comprehensive queries

set -e

cd "$(dirname "$0")/terraform"

REGION="us-west-2"
CLUSTER_NAME=$(terraform output -raw ecs_cluster_name | tr -d '\n\r')
SERVICE_NAME=$(terraform output -raw ecs_service_name | tr -d '\n\r')
RDS_ENDPOINT=$(terraform output -raw rds_endpoint | tr -d '\n\r')
RDS_USER=$(terraform output -raw rds_username | tr -d '\n\r')
RDS_DB=$(terraform output -raw rds_database_name | tr -d '\n\r')
RDS_PWD=$(terraform output -raw rds_password | tr -d '\n\r')
# Get subnet IDs and SG from existing ECS service (most reliable)
SUBNET_IDS=""
SG_ID=""

# Get execution role (needed for Fargate tasks)
EXECUTION_ROLE=$(aws iam list-roles --query 'Roles[?RoleName==`LabRole`].Arn' --output text 2>/dev/null || echo "")

if [ -z "$EXECUTION_ROLE" ]; then
  echo "⚠ Could not find LabRole. Trying to get from data source..."
  EXECUTION_ROLE=$(terraform output -raw execution_role_arn 2>/dev/null || echo "")
fi

if [ -z "$EXECUTION_ROLE" ]; then
  echo "Error: Could not find execution role. Please ensure LabRole exists."
  exit 1
fi

OUTPUT_DIR="../reports/HW8"
mkdir -p "$OUTPUT_DIR"
QUERY_RESULTS_FILE="$OUTPUT_DIR/db_verification_results.txt"

echo "=========================================="
echo "RDS Query via ECS Task"
echo "=========================================="
echo "Cluster: $CLUSTER_NAME"
echo "RDS Endpoint: $RDS_ENDPOINT"
echo "Database: $RDS_DB"
echo "Results will be saved to: $QUERY_RESULTS_FILE"
echo ""

# Create comprehensive SQL query (from query_rds.sh)
QUERY_SQL=$(cat << 'SQL_EOF'
SELECT "=== TABLES ===" AS info;
SHOW TABLES;

SELECT "\n=== shopping_carts SCHEMA ===" AS info;
DESCRIBE shopping_carts;

SELECT "\n=== shopping_cart_items SCHEMA ===" AS info;
DESCRIBE shopping_cart_items;

SELECT "\n=== INDEXES ===" AS info;
SHOW INDEX FROM shopping_carts;
SHOW INDEX FROM shopping_cart_items;

SELECT "\n=== ROW COUNTS ===" AS info;
SELECT 
  (SELECT COUNT(*) FROM shopping_carts) AS total_carts,
  (SELECT COUNT(*) FROM shopping_cart_items) AS total_items;

SELECT "\n=== SAMPLE CARTS (first 5) ===" AS info;
SELECT * FROM shopping_carts LIMIT 5;

SELECT "\n=== SAMPLE ITEMS (first 10) ===" AS info;
SELECT * FROM shopping_cart_items LIMIT 10;

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
SQL_EOF
)

LOG_GROUP="/ecs/rds-query-temp"
LOG_STREAM="query-$(date +%s)"

# Ensure log group exists
aws logs create-log-group --log-group-name "$LOG_GROUP" --region "$REGION" 2>/dev/null || true

echo "Creating ECS task definition..."
cat > /tmp/query-task.json <<EOF
{
  "family": "rds-query-task-$(date +%s)",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "executionRoleArn": "$EXECUTION_ROLE",
  "taskRoleArn": "$EXECUTION_ROLE",
  "containerDefinitions": [
    {
      "name": "mysql-query",
      "image": "mysql:8.0",
      "essential": true,
      "command": [
        "sh",
        "-c",
        "mysql -h $RDS_ENDPOINT -u $RDS_USER -p'$RDS_PWD' $RDS_DB -e \"SHOW TABLES; DESCRIBE shopping_carts; DESCRIBE shopping_cart_items; SHOW INDEX FROM shopping_carts; SHOW INDEX FROM shopping_cart_items; SELECT (SELECT COUNT(*) FROM shopping_carts) AS total_carts, (SELECT COUNT(*) FROM shopping_cart_items) AS total_items; SELECT * FROM shopping_carts LIMIT 5; SELECT * FROM shopping_cart_items LIMIT 10; SELECT c.shopping_cart_id, c.customer_id, c.status, COUNT(i.product_id) AS item_count, SUM(i.quantity) AS total_quantity, c.created_at FROM shopping_carts c LEFT JOIN shopping_cart_items i ON c.shopping_cart_id = i.shopping_cart_id GROUP BY c.shopping_cart_id, c.customer_id, c.status, c.created_at ORDER BY c.shopping_cart_id DESC LIMIT 10;\""
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "$LOG_GROUP",
          "awslogs-region": "$REGION",
          "awslogs-stream-prefix": "query"
        }
      }
    }
  ]
}
EOF

# Register task definition
TASK_DEF_ARN=$(aws ecs register-task-definition \
  --cli-input-json file:///tmp/query-task.json \
  --region "$REGION" \
  --query 'taskDefinition.taskDefinitionArn' \
  --output text)

echo "✓ Task definition registered: $TASK_DEF_ARN"
echo ""

# Get subnet IDs and security group
# Subnets are in the network module, so query from AWS or get from existing ECS service
if [ -z "$SUBNET_IDS" ]; then
  # Try to get from existing ECS service
  SUBNET_IDS=$(aws ecs describe-services \
    --cluster "$CLUSTER_NAME" \
    --services "$SERVICE_NAME" \
    --region "$REGION" \
    --query 'services[0].networkConfiguration.awsvpcConfiguration.subnets[]' \
    --output text 2>/dev/null || echo "")
  
  # Fallback: get default VPC subnets
  if [ -z "$SUBNET_IDS" ]; then
    VPC_ID=$(aws ec2 describe-vpcs --filters Name=isDefault,Values=true --query 'Vpcs[0].VpcId' --output text --region "$REGION")
    SUBNET_IDS=$(aws ec2 describe-subnets \
      --filters "Name=vpc-id,Values=$VPC_ID" \
      --query 'Subnets[*].SubnetId' \
      --output text \
      --region "$REGION" | awk '{print $1" "$2}')
  fi
fi

if [ -z "$SG_ID" ]; then
  # Try to get from existing ECS service
  SG_ID=$(aws ecs describe-services \
    --cluster "$CLUSTER_NAME" \
    --services "$SERVICE_NAME" \
    --region "$REGION" \
    --query 'services[0].networkConfiguration.awsvpcConfiguration.securityGroups[0]' \
    --output text 2>/dev/null || echo "")
  
  # Fallback: search by name pattern
  if [ -z "$SG_ID" ]; then
    SG_ID=$(aws ec2 describe-security-groups \
      --filters "Name=tag:Name,Values=*${CLUSTER_NAME%-cluster}*" "Name=description,Values=*ECS*" \
      --query 'SecurityGroups[0].GroupId' \
      --output text \
      --region "$REGION" 2>/dev/null || echo "")
  fi
fi

# Convert subnet IDs (tab or space-separated) to array and then to comma-separated for JSON
SUBNET_LIST=$(echo "$SUBNET_IDS" | tr '\t' ' ' | tr -s ' ' | tr ' ' ',')

echo "Running ECS task..."
echo "  Subnets: $SUBNET_LIST"
echo "  Security Group: $SG_ID"
echo ""

# Build network config JSON (properly format subnet IDs as array)
SUBNET_ARRAY=$(echo "$SUBNET_LIST" | sed 's/,/","/g' | sed 's/^/"/' | sed 's/$/"/')
NETWORK_CONFIG="{\"awsvpcConfiguration\":{\"subnets\":[$SUBNET_ARRAY],\"securityGroups\":[\"$SG_ID\"],\"assignPublicIp\":\"ENABLED\"}}"

RUN_TASK_OUTPUT=$(aws ecs run-task \
  --cluster "$CLUSTER_NAME" \
  --task-definition "$TASK_DEF_ARN" \
  --launch-type FARGATE \
  --network-configuration "$NETWORK_CONFIG" \
  --region "$REGION")

TASK_ARN=$(echo "$RUN_TASK_OUTPUT" | jq -r '.tasks[0].taskArn // empty')

if [ -z "$TASK_ARN" ]; then
  echo "❌ Failed to start task. Output:"
  echo "$RUN_TASK_OUTPUT" | jq '.'
  exit 1
fi

echo "✓ Task started: $TASK_ARN"
echo ""
echo "Waiting for task to complete (this may take 30-60 seconds)..."
echo ""

# Wait for task to stop (max 2 minutes)
TIMEOUT=120
ELAPSED=0
while [ $ELAPSED -lt $TIMEOUT ]; do
  TASK_STATUS=$(aws ecs describe-tasks \
    --cluster "$CLUSTER_NAME" \
    --tasks "$TASK_ARN" \
    --region "$REGION" \
    --query 'tasks[0].lastStatus' \
    --output text)
  
  if [ "$TASK_STATUS" = "STOPPED" ]; then
    break
  fi
  
  sleep 5
  ELAPSED=$((ELAPSED + 5))
  echo -n "."
done

echo ""
echo ""

# Get exit code
EXIT_CODE=$(aws ecs describe-tasks \
  --cluster "$CLUSTER_NAME" \
  --tasks "$TASK_ARN" \
  --region "$REGION" \
  --query 'tasks[0].containers[0].exitCode' \
  --output text)

if [ "$EXIT_CODE" != "0" ] && [ "$EXIT_CODE" != "null" ]; then
  echo "⚠ Task completed with exit code: $EXIT_CODE"
  echo "Checking logs for errors..."
  echo ""
fi

# Fetch logs
echo "Fetching query results from CloudWatch Logs..."
echo ""

LOG_STREAM_NAME=$(aws logs describe-log-streams \
  --log-group-name "$LOG_GROUP" \
  --order-by LastEventTime \
  --descending \
  --max-items 1 \
  --region "$REGION" \
  --query 'logStreams[0].logStreamName' \
  --output text 2>/dev/null || echo "")

if [ -n "$LOG_STREAM_NAME" ] && [ "$LOG_STREAM_NAME" != "null" ]; then
  echo "=========================================="
  echo "Query Results"
  echo "=========================================="
  echo ""
  
  # Get and display results
  QUERY_OUTPUT=$(aws logs get-log-events \
    --log-group-name "$LOG_GROUP" \
    --log-stream-name "$LOG_STREAM_NAME" \
    --region "$REGION" \
    --query 'events[*].message' \
    --output text 2>/dev/null | sed 's/\\t/\t/g' | sed 's/\\n/\n/g')
  
  echo "$QUERY_OUTPUT"
  echo ""
  echo "=========================================="
  
  # Save to file
  {
    echo "Database Verification Results"
    echo "Generated: $(date -u +"%Y-%m-%d %H:%M:%S UTC")"
    echo "RDS Endpoint: $RDS_ENDPOINT"
    echo "Database: $RDS_DB"
    echo "=========================================="
    echo ""
    echo "$QUERY_OUTPUT"
  } > "$QUERY_RESULTS_FILE"
  
  echo ""
  echo "✅ Results saved to: $QUERY_RESULTS_FILE"
else
  echo "⚠ Could not retrieve logs. Task may still be running or logs not available yet."
  echo "Check manually with:"
  echo "  aws logs tail $LOG_GROUP --follow --region $REGION"
fi

# Cleanup task definition (optional - comment out if you want to keep it)
echo ""
echo "Cleaning up task definition..."
aws ecs deregister-task-definition \
  --task-definition "$TASK_DEF_ARN" \
  --region "$REGION" > /dev/null 2>&1 || true

echo "✓ Done!"
echo ""
echo "Note: Log group '$LOG_GROUP' was created. You can delete it with:"
echo "  aws logs delete-log-group --log-group-name $LOG_GROUP --region $REGION"

