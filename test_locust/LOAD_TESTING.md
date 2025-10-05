# Load Testing Guide for Product API

## Locust Configuration Parameters

In the Locust Web UI, you'll configure these three parameters:

1. **Number of users (peak concurrency)**: Total number of simulated users
2. **Ramp up (users started/second)**: How fast to spawn users
3. **Run time**: Duration of the test (e.g., 5m, 10s, 1h)

## Test Scenarios

### Scenario 1: Normal Load (Browsing Heavy)
Simulates typical e-commerce traffic where users browse more than they add products.

**Locust Settings:**
- **Number of users**: 100
- **Ramp up**: 10 users/sec
- **Run time**: 5m

**Expected Results**: 
- GET requests: ~71% of traffic (task weight: 10/14)
- POST requests: ~21% of traffic (task weight: 3/14)
- Health checks: ~7% of traffic (task weight: 1/14)
- **Total task weight**: 14 (10 + 3 + 1)
- Target: < 200ms response time for GET, < 500ms for POST

### Scenario 2: Write-Heavy Load (Inventory Update)
Simulates bulk product updates (e.g., price changes, inventory sync).

**Locust Settings:**
- **Number of users**: 50
- **Ramp up**: 5 users/sec
- **Run time**: 3m

**Note**: Modify task weights in code - increase POST to weight 8, reduce GET to weight 2

### Scenario 3: Spike Test
Simulates sudden traffic spike (e.g., flash sale announcement).

**Locust Settings:**
- **Number of users**: 500
- **Ramp up**: 25 users/sec (reaches 500 users in ~20 seconds)
- **Run time**: 3m

**Expected**: System should handle the spike gracefully

## HttpUser vs FastHttpUser Comparison

### Available Test Classes

**ProductAPIUser (HttpUser)**
- Uses Python `requests` library
- Better error handling and response validation
- Slower but more reliable
- Good for functional testing

**FastProductAPIUser (FastHttpUser)**
- Uses HTTP/1.1 keep-alive connections
- Connection pooling for better performance  
- Less overhead, higher throughput
- Better for performance testing

### Running Comparisons

```bash
# Test with both classes (default)
/Users/haoniu/Desktop/CS6650_Online_Store/test_locust/venv/bin/locust -f simple_locustfile.py --host=http://localhost:8080

# Test with HttpUser only (better validation)
/Users/haoniu/Desktop/CS6650_Online_Store/test_locust/venv/bin/locust -f simple_locustfile.py --host=http://localhost:8080 ProductAPIUser

# Test with FastHttpUser only (better performance)
/Users/haoniu/Desktop/CS6650_Online_Store/test_locust/venv/bin/locust -f simple_locustfile.py --host=http://localhost:8080 FastProductAPIUser
```

**Expected Performance Difference**:
- FastHttpUser: ~20-30% higher RPS
- HttpUser: Better error reporting and validation

## Metrics to Watch

### Key Performance Indicators (KPIs)
- **Response Time (p50, p95, p99)**: Should be < 200ms for GET, < 500ms for POST
- **Requests per Second (RPS)**: Target 1000+ RPS
- **Error Rate**: Should be < 1% for valid requests  
- **Throughput**: MB/s transferred

### AWS CloudWatch Metrics (when testing deployed API)
- ECS CPU utilization (should stay < 80%)
- ECS Memory utilization (should stay < 80%)
- ALB target response time
- ALB 4XX and 5XX errors

## Running Tests

### Local Testing
```bash
# Start your Go server
go run cmd/server/main.go

# In another terminal, start Locust
cd test_locust
/Users/haoniu/Desktop/CS6650_Online_Store/test_locust/venv/bin/locust -f simple_locustfile.py --host=http://localhost:8080

# Open browser
open http://localhost:8089
```

### Testing Deployed API
```bash
# Test against your AWS ECS deployment
/Users/haoniu/Desktop/CS6650_Online_Store/test_locust/venv/bin/locust -f simple_locustfile.py --host=http://35.166.6.135:8080