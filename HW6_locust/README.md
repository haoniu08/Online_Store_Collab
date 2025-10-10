# Homework 6 Load Testing Setup

This directory contains the load testing configuration specifically for Homework 6 requirements.

## Test Scenarios

### Test 1 - Baseline (5 users, 2 minutes)
```bash
./run_tests.sh baseline
```
- **Expected**: Moderate CPU (~60%), fast responses (<50ms)
- **Purpose**: Establish performance baseline

### Test 2 - Breaking Point (20 users, 3 minutes)
```bash
./run_tests.sh breaking
```
- **Expected**: High CPU (~100%), degraded responses (>200ms)
- **Purpose**: Find the system's breaking point

### Comprehensive Test (10 users, 5 minutes)
```bash
./run_tests.sh comprehensive
```
- **Purpose**: Validate search functionality and edge cases

## Quick Start

### Local Testing
1. Start your server:
   ```bash
   cd .. && ./server
   ```

2. Run tests:
   ```bash
   chmod +x run_tests.sh
   ./run_tests.sh all
   ```

3. View results in `./results/` directory

### Docker Testing
1. Start the full stack:
   ```bash
   docker-compose up -d
   ```

2. Access Locust Web UI: http://localhost:8089
3. Monitor container resources: http://localhost:8081 (cAdvisor)

### AWS ECS Testing
```bash
HOST=http://your-alb-dns-name ./run_tests.sh all
```

## Files

- `locustfile.py` - Main load testing scenarios
- `run_tests.sh` - Test execution script
- `docker-compose.yml` - Docker setup with resource limits
- `results/` - Test reports and data (created automatically)

## Key Metrics to Monitor

### Local/Docker
- CPU usage (should hit ~100% with 20 users)
- Memory usage (should stay stable)
- Response times (should degrade significantly)

### AWS ECS
- CloudWatch CPU Utilization
- CloudWatch Memory Utilization
- ECS Service metrics
- ALB target response time (if using Part III setup)

## Expected Results

With 256 CPU units / 512MB memory configuration:
- **5 users**: System should handle easily
- **20 users**: System should show stress and degraded performance
- **Resource bottleneck**: CPU should be the limiting factor

This demonstrates the need for horizontal scaling (Part III of Homework 6).