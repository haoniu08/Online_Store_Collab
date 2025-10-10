#!/bin/bash

# Homework 6 Load Testing Scripts
# ===============================
# 
# These scripts run the specific test scenarios required by Homework 6:
# Test 1 - Baseline: 5 users for 2 minutes
# Test 2 - Breaking Point: 20 users for 3 minutes

set -e

echo "🚀 Homework 6 Load Testing Suite"
echo "================================="

# Configuration
HOST="${HOST:-http://localhost:8080}"
LOCUST_FILE="locustfile.py"

# Try to find locust in virtual environment or system
LOCUST_CMD=""
if [ -f "../test_locust/venv/bin/locust" ]; then
    LOCUST_CMD="../test_locust/venv/bin/locust"
elif command -v locust >/dev/null 2>&1; then
    LOCUST_CMD="locust"
else
    echo -e "${RED}❌ Locust not found!${NC}"
    echo "Please install locust:"
    echo "  pip3 install locust"
    echo "Or use Docker:"
    echo "  docker-compose up"
    exit 1
fi

echo -e "${BLUE}📍 Using Locust: $LOCUST_CMD${NC}"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to check if server is running
check_server() {
    echo -e "${BLUE}🔍 Checking if server is running on $HOST...${NC}"
    if curl -s -f "$HOST/health" > /dev/null; then
        echo -e "${GREEN}✅ Server is running and healthy${NC}"
        return 0
    else
        echo -e "${RED}❌ Server is not running on $HOST${NC}"
        echo ""
        echo "Please start your server first:"
        echo "  cd /Users/haoniu/Desktop/CS6650_Online_Store"
        echo "  ./server"
        echo ""
        echo "Or if using Docker:"
        echo "  docker-compose up product-api"
        exit 1
    fi
}

# Function to create results directory
setup_results() {
    mkdir -p results
    echo -e "${BLUE}📁 Results will be saved to ./results/${NC}"
}

# Function to run baseline test
run_baseline_test() {
    echo -e "\n${YELLOW}📊 Running Test 1 - Baseline (5 users, 2 minutes)${NC}"
    echo "=================================================="
    echo "Expected behavior with 256 CPU/512MB memory:"
    echo "- Moderate CPU usage (~60%)"
    echo "- Fast response times (<50ms)"
    echo "- Steady memory usage"
    echo ""
    
    $LOCUST_CMD \
        -f "$LOCUST_FILE" \
        --host="$HOST" \
        --users=5 \
        --spawn-rate=1 \
        --run-time=2m \
        --headless \
        --html=results/baseline_test_report.html \
        --csv=results/baseline_test \
        --logfile=results/baseline_test.log
    
    echo -e "${GREEN}✅ Baseline test completed${NC}"
    echo "📊 Report: results/baseline_test_report.html"
    echo "📈 CSV data: results/baseline_test_*.csv"
    echo "📝 Logs: results/baseline_test.log"
}

# Function to run breaking point test
run_breaking_point_test() {
    echo -e "\n${YELLOW}💥 Running Test 2 - Breaking Point (20 users, 3 minutes)${NC}"
    echo "========================================================"
    echo "Expected behavior with 256 CPU/512MB memory:"
    echo "- High CPU usage (near 100%)"
    echo "- Degraded response times (>200ms)"
    echo "- System under stress"
    echo "- This should break the system!"
    echo ""
    
    $LOCUST_CMD \
        -f "$LOCUST_FILE" \
        --host="$HOST" \
        --users=20 \
        --spawn-rate=2 \
        --run-time=3m \
        --headless \
        --html=results/breaking_point_test_report.html \
        --csv=results/breaking_point_test \
        --logfile=results/breaking_point_test.log
    
    echo -e "${GREEN}✅ Breaking point test completed${NC}"
    echo "📊 Report: results/breaking_point_test_report.html"
    echo "📈 CSV data: results/breaking_point_test_*.csv"
    echo "📝 Logs: results/breaking_point_test.log"
}

# Function to run comprehensive test
run_comprehensive_test() {
    echo -e "\n${YELLOW}🔬 Running Comprehensive Search Test (10 users, 5 minutes)${NC}"
    echo "==========================================================="
    echo "This test validates all search patterns and responses"
    echo ""
    
    $LOCUST_CMD \
        -f "$LOCUST_FILE" \
        --host="$HOST" \
        --users=10 \
        --spawn-rate=2 \
        --run-time=5m \
        --headless \
        --html=results/comprehensive_test_report.html \
        --csv=results/comprehensive_test \
        --logfile=results/comprehensive_test.log
    
    echo -e "${GREEN}✅ Comprehensive test completed${NC}"
    echo "📊 Report: results/comprehensive_test_report.html"
    echo "📈 CSV data: results/comprehensive_test_*.csv"
    echo "📝 Logs: results/comprehensive_test.log"
}

# Function to show test summary
show_summary() {
    echo -e "\n${GREEN}🎉 Load Testing Summary${NC}"
    echo "======================="
    
    if [ -f "results/baseline_test_stats.csv" ] && [ -f "results/breaking_point_test_stats.csv" ]; then
        echo "Comparing baseline vs breaking point:"
        echo ""
        echo "Baseline Test (5 users):"
        tail -n 1 results/baseline_test_stats.csv | awk -F',' '{printf "  Average Response Time: %s ms\n  RPS: %s\n", $9, $10}' || echo "  Data processing error"
        
        echo ""
        echo "Breaking Point Test (20 users):"
        tail -n 1 results/breaking_point_test_stats.csv | awk -F',' '{printf "  Average Response Time: %s ms\n  RPS: %s\n", $9, $10}' || echo "  Data processing error"
    fi
    
    echo ""
    echo "📁 All results saved in ./results/"
    echo "📊 Open HTML reports in your browser for detailed analysis"
    echo ""
    echo "Next steps for Homework 6:"
    echo "1. 📈 Check CloudWatch metrics in AWS Console"
    echo "2. 📝 Document response time degradation"
    echo "3. 🔍 Identify which resource (CPU/Memory) hit limits first"
    echo "4. 📋 Write your analysis report"
}

# Main execution
main() {
    setup_results
    check_server
    
    case "${1:-all}" in
        "baseline")
            run_baseline_test
            ;;
        "breaking")
            run_breaking_point_test
            ;;
        "comprehensive")
            run_comprehensive_test
            ;;
        "all")
            echo -e "${BLUE}🚀 Running all Homework 6 test scenarios...${NC}"
            run_baseline_test
            echo -e "\n${YELLOW}⏳ Cooling down for 30 seconds...${NC}"
            sleep 30
            run_breaking_point_test
            echo -e "\n${YELLOW}⏳ Final cooldown...${NC}"
            sleep 30
            run_comprehensive_test
            show_summary
            ;;
        *)
            echo "Usage: $0 [baseline|breaking|comprehensive|all]"
            echo ""
            echo "Homework 6 Test Scenarios:"
            echo "  baseline      - Test 1: 5 users, 2 minutes"
            echo "  breaking      - Test 2: 20 users, 3 minutes" 
            echo "  comprehensive - Validation: 10 users, 5 minutes"
            echo "  all           - Run all tests sequentially (default)"
            echo ""
            echo "Environment Variables:"
            echo "  HOST          - Target server (default: http://localhost:8080)"
            echo ""
            echo "Examples:"
            echo "  $0 baseline                    # Run just baseline test"
            echo "  HOST=http://your-alb-url $0    # Test against AWS ALB"
            exit 1
            ;;
    esac
    
    if [ "$1" != "all" ]; then
        show_summary
    fi
}

main "$@"