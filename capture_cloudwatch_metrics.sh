#!/bin/bash
# CloudWatch Metrics Capture Script for HW8 Part 6
# Pulls RDS, ECS, and ALB metrics and saves to reports/HW8/

set -e

cd "$(dirname "$0")/terraform"

REGION="us-west-2"
RDS_ID=$(terraform output -raw rds_endpoint | cut -d. -f1 | tr -d '\n\r')
CLUSTER_NAME=$(terraform output -raw ecs_cluster_name | tr -d '\n\r')
SERVICE_NAME=$(terraform output -raw ecs_service_name | tr -d '\n\r')
ALB_NAME=$(terraform output -raw alb_dns_name | cut -d. -f1 | tr -d '\n\r')

OUTPUT_DIR="../reports/HW8/metrics"
mkdir -p "$OUTPUT_DIR"

# Time range: last 6 hours (adjust as needed)
END_TIME=$(date -u +"%Y-%m-%dT%H:%M:%S")
START_TIME=$(date -u -v-6H +"%Y-%m-%dT%H:%M:%S" 2>/dev/null || date -u -d "6 hours ago" +"%Y-%m-%dT%H:%M:%S")

echo "=========================================="
echo "CloudWatch Metrics Capture"
echo "=========================================="
echo "Time Range: $START_TIME to $END_TIME (UTC)"
echo "Region: $REGION"
echo "Output Directory: $OUTPUT_DIR"
echo ""

# RDS Metrics
echo "📊 Capturing RDS Metrics..."
echo "  - CPU Utilization"
aws cloudwatch get-metric-statistics \
  --namespace AWS/RDS \
  --metric-name CPUUtilization \
  --dimensions Name=DBInstanceIdentifier,Value="$RDS_ID" \
  --start-time "$START_TIME" \
  --end-time "$END_TIME" \
  --period 300 \
  --statistics Average --statistics Maximum \
  --region "$REGION" \
  --output json > "$OUTPUT_DIR/rds_cpu.json"

echo "  - Database Connections"
aws cloudwatch get-metric-statistics \
  --namespace AWS/RDS \
  --metric-name DatabaseConnections \
  --dimensions Name=DBInstanceIdentifier,Value="$RDS_ID" \
  --start-time "$START_TIME" \
  --end-time "$END_TIME" \
  --period 300 \
  --statistics Average --statistics Maximum --statistics Minimum \
  --region "$REGION" \
  --output json > "$OUTPUT_DIR/rds_connections.json"

echo "  - Read IOPS"
aws cloudwatch get-metric-statistics \
  --namespace AWS/RDS \
  --metric-name ReadIOPS \
  --dimensions Name=DBInstanceIdentifier,Value="$RDS_ID" \
  --start-time "$START_TIME" \
  --end-time "$END_TIME" \
  --period 300 \
  --statistics Average --statistics Maximum \
  --region "$REGION" \
  --output json > "$OUTPUT_DIR/rds_read_iops.json"

echo "  - Write IOPS"
aws cloudwatch get-metric-statistics \
  --namespace AWS/RDS \
  --metric-name WriteIOPS \
  --dimensions Name=DBInstanceIdentifier,Value="$RDS_ID" \
  --start-time "$START_TIME" \
  --end-time "$END_TIME" \
  --period 300 \
  --statistics Average --statistics Maximum \
  --region "$REGION" \
  --output json > "$OUTPUT_DIR/rds_write_iops.json"

echo "  - Freeable Memory"
aws cloudwatch get-metric-statistics \
  --namespace AWS/RDS \
  --metric-name FreeableMemory \
  --dimensions Name=DBInstanceIdentifier,Value="$RDS_ID" \
  --start-time "$START_TIME" \
  --end-time "$END_TIME" \
  --period 300 \
  --statistics Average --statistics Minimum \
  --region "$REGION" \
  --output json > "$OUTPUT_DIR/rds_memory.json"

# ECS Metrics
echo ""
echo "📊 Capturing ECS Metrics..."
echo "  - CPU Utilization"
aws cloudwatch get-metric-statistics \
  --namespace AWS/ECS \
  --metric-name CPUUtilization \
  --dimensions Name=ClusterName,Value="$CLUSTER_NAME" Name=ServiceName,Value="$SERVICE_NAME" \
  --start-time "$START_TIME" \
  --end-time "$END_TIME" \
  --period 300 \
  --statistics Average --statistics Maximum \
  --region "$REGION" \
  --output json > "$OUTPUT_DIR/ecs_cpu.json"

echo "  - Memory Utilization"
aws cloudwatch get-metric-statistics \
  --namespace AWS/ECS \
  --metric-name MemoryUtilization \
  --dimensions Name=ClusterName,Value="$CLUSTER_NAME" Name=ServiceName,Value="$SERVICE_NAME" \
  --start-time "$START_TIME" \
  --end-time "$END_TIME" \
  --period 300 \
  --statistics Average --statistics Maximum \
  --region "$REGION" \
  --output json > "$OUTPUT_DIR/ecs_memory.json"

# ALB Metrics
echo ""
echo "📊 Capturing ALB Metrics..."
echo "  - Target Response Time"
aws cloudwatch get-metric-statistics \
  --namespace AWS/ApplicationELB \
  --metric-name TargetResponseTime \
  --dimensions Name=LoadBalancer,Value="app/$ALB_NAME/*" \
  --start-time "$START_TIME" \
  --end-time "$END_TIME" \
  --period 300 \
  --statistics Average --statistics Maximum \
  --region "$REGION" \
  --output json > "$OUTPUT_DIR/alb_response_time.json"

echo "  - Request Count"
aws cloudwatch get-metric-statistics \
  --namespace AWS/ApplicationELB \
  --metric-name RequestCount \
  --dimensions Name=LoadBalancer,Value="app/$ALB_NAME/*" \
  --start-time "$START_TIME" \
  --end-time "$END_TIME" \
  --period 300 \
  --statistics Sum \
  --region "$REGION" \
  --output json > "$OUTPUT_DIR/alb_request_count.json"

echo "  - HTTP 5xx Errors"
aws cloudwatch get-metric-statistics \
  --namespace AWS/ApplicationELB \
  --metric-name HTTPCode_Target_5XX_Count \
  --dimensions Name=LoadBalancer,Value="app/$ALB_NAME/*" \
  --start-time "$START_TIME" \
  --end-time "$END_TIME" \
  --period 300 \
  --statistics Sum \
  --region "$REGION" \
  --output json > "$OUTPUT_DIR/alb_5xx_errors.json"

# Generate summary report
echo ""
echo "📝 Generating Summary Report..."
cat > "$OUTPUT_DIR/metrics_summary.txt" << EOF
CloudWatch Metrics Summary
Generated: $(date -u +"%Y-%m-%d %H:%M:%S UTC")
Time Range: $START_TIME to $END_TIME (UTC)
Region: $REGION

RDS Metrics:
- CPU Utilization: $OUTPUT_DIR/rds_cpu.json
- Database Connections: $OUTPUT_DIR/rds_connections.json
- Read IOPS: $OUTPUT_DIR/rds_read_iops.json
- Write IOPS: $OUTPUT_DIR/rds_write_iops.json
- Freeable Memory: $OUTPUT_DIR/rds_memory.json

ECS Metrics:
- CPU Utilization: $OUTPUT_DIR/ecs_cpu.json
- Memory Utilization: $OUTPUT_DIR/ecs_memory.json

ALB Metrics:
- Target Response Time: $OUTPUT_DIR/alb_response_time.json
- Request Count: $OUTPUT_DIR/alb_request_count.json
- HTTP 5xx Errors: $OUTPUT_DIR/alb_5xx_errors.json

To view metrics in readable format:
  cat $OUTPUT_DIR/rds_cpu.json | jq '.Datapoints[] | {Timestamp, Average, Maximum}'

EOF

echo "✅ Metrics captured successfully!"
echo ""
echo "Files saved to: $OUTPUT_DIR"
echo ""
echo "Quick stats (last 6 hours):"
echo ""

# Show quick stats if jq is available
if command -v jq &> /dev/null; then
  echo "RDS CPU (avg/max):"
  jq -r '.Datapoints | if length > 0 then map(.Average) | add/length | "  Average: \(.)%" else "  No data" end' "$OUTPUT_DIR/rds_cpu.json" 2>/dev/null || echo "  Processing..."
  
  echo "RDS Connections (avg/max):"
  jq -r '.Datapoints | if length > 0 then "  Average: \(map(.Average) | add/length | floor), Max: \(map(.Maximum) | max | floor)" else "  No data" end' "$OUTPUT_DIR/rds_connections.json" 2>/dev/null || echo "  Processing..."
  
  echo "ALB Response Time (avg/max ms):"
  jq -r '.Datapoints | if length > 0 then map(.Average) | add/length * 1000 | "  Average: \(.)ms" else "  No data" end' "$OUTPUT_DIR/alb_response_time.json" 2>/dev/null || echo "  Processing..."
else
  echo "  (Install 'jq' for formatted output: brew install jq)"
fi

echo ""
echo "To view full details, check JSON files in: $OUTPUT_DIR"

