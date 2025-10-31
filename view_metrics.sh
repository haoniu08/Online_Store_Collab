#!/bin/bash
# Helper script to view CloudWatch metrics in readable format

METRICS_DIR="reports/HW8/metrics"

if [ ! -d "$METRICS_DIR" ]; then
  echo "Error: Metrics directory not found. Run capture_cloudwatch_metrics.sh first."
  exit 1
fi

echo "=========================================="
echo "CloudWatch Metrics Viewer"
echo "=========================================="
echo ""

if ! command -v jq &> /dev/null; then
  echo "⚠ jq is required for this script. Install with: brew install jq"
  exit 1
fi

echo "📊 RDS CPU Utilization (%):"
echo "------------------------"
jq -r '.Datapoints | sort_by(.Timestamp) | .[] | "\(.Timestamp) | Avg: \(.Average // "N/A")% | Max: \(.Maximum // "N/A")%"' "$METRICS_DIR/rds_cpu.json" 2>/dev/null | head -10
echo ""

echo "📊 RDS Database Connections:"
echo "------------------------"
jq -r '.Datapoints | sort_by(.Timestamp) | .[] | "\(.Timestamp) | Avg: \(.Average // "N/A") | Max: \(.Maximum // "N/A") | Min: \(.Minimum // "N/A")"' "$METRICS_DIR/rds_connections.json" 2>/dev/null | head -10
echo ""

echo "📊 RDS Read IOPS:"
echo "------------------------"
jq -r '.Datapoints | sort_by(.Timestamp) | .[] | "\(.Timestamp) | Avg: \(.Average // "N/A") | Max: \(.Maximum // "N/A")"' "$METRICS_DIR/rds_read_iops.json" 2>/dev/null | head -10
echo ""

echo "📊 RDS Write IOPS:"
echo "------------------------"
jq -r '.Datapoints | sort_by(.Timestamp) | .[] | "\(.Timestamp) | Avg: \(.Average // "N/A") | Max: \(.Maximum // "N/A")"' "$METRICS_DIR/rds_write_iops.json" 2>/dev/null | head -10
echo ""

echo "📊 ALB Response Time (ms):"
echo "------------------------"
jq -r '.Datapoints | sort_by(.Timestamp) | .[] | "\(.Timestamp) | Avg: \((.Average // 0) * 1000)ms | Max: \((.Maximum // 0) * 1000)ms"' "$METRICS_DIR/alb_response_time.json" 2>/dev/null | head -10
echo ""

echo "📊 ALB Request Count:"
echo "------------------------"
jq -r '.Datapoints | sort_by(.Timestamp) | .[] | "\(.Timestamp) | Total: \(.Sum // "N/A")"' "$METRICS_DIR/alb_request_count.json" 2>/dev/null | head -10
echo ""

echo "📊 ECS CPU Utilization (%):"
echo "------------------------"
jq -r '.Datapoints | sort_by(.Timestamp) | .[] | "\(.Timestamp) | Avg: \(.Average // "N/A")% | Max: \(.Maximum // "N/A")%"' "$METRICS_DIR/ecs_cpu.json" 2>/dev/null | head -10
echo ""

echo "Full JSON files available in: $METRICS_DIR"

