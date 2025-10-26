# CS6650 Online Store Collaboration Project

## Overview

This project demonstrates the evolution of an e-commerce order processing system from a simple synchronous approach to a sophisticated serverless architecture. We built and tested three different approaches to handle flash sale traffic, learning when to optimize code versus when to scale infrastructure.

**What We Built:** An online store that can handle massive traffic spikes (like Black Friday sales) without crashing or making customers wait.

**Why It Matters:** Real e-commerce sites face sudden traffic spikes that can break traditional systems. We show how modern cloud architecture solves these problems.

---

## Project Structure

```
Online_Store_Collab/
├── cmd/                          # Application entry points
│   ├── server/                   # Main API server (orders, products)
│   ├── processor/                # Background order processor (ECS)
│   ├── lambda/                   # Serverless order processor
│   └── test/                     # Testing utilities
├── internal/                     # Core business logic
│   ├── handlers/                 # HTTP request handlers
│   ├── models/                   # Data structures (Order, Product)
│   ├── store/                    # Data storage logic
│   └── worker/                   # Order processing logic
├── terraform/                    # Infrastructure as Code
│   ├── modules/                  # Reusable infrastructure components
│   │   ├── alb/                  # Load balancer
│   │   ├── ecs/                  # Container orchestration
│   │   ├── lambda/               # Serverless functions
│   │   ├── network/              # VPC and security
│   │   ├── sns/                  # Message publishing
│   │   └── sqs/                  # Message queuing
│   └── main.tf                   # Main infrastructure configuration
├── test_locust/                  # Load testing scripts
├── HW6_locust/                   # Homework 6 load tests
└── Dockerfile*                   # Container configurations
```

---

## Architecture Evolution

### Phase 1: Synchronous Processing (The Problem)
```
Customer → API → Payment Processing (3 seconds) → Response
```
**Problem:** When 20 customers try to order simultaneously, they all wait 3 seconds each, creating a bottleneck.

### Phase 2: Asynchronous Processing (The Solution)
```
Customer → API → Queue → Immediate Response (< 100ms)
                ↓
         Background Workers → Payment Processing
```
**Solution:** Accept orders instantly, process them in the background using AWS SNS/SQS.

### Phase 3: Serverless Processing (The Optimization)
```
Customer → API → SNS → Lambda → Automatic Scaling
```
**Optimization:** Eliminate infrastructure management entirely using AWS Lambda.

---

## Key Components Explained

### 🏪 **Order Processing System**
- **What it does:** Handles customer orders with payment verification
- **The challenge:** Payment processing takes 3 seconds (simulating real payment gateways)
- **The problem:** Multiple customers ordering simultaneously creates delays

### 🔄 **Message Queuing (SNS/SQS)**
- **SNS (Simple Notification Service):** Like a megaphone that broadcasts order events
- **SQS (Simple Queue Service):** Like a waiting line that holds orders until workers can process them
- **Why we need it:** Decouples order acceptance from order processing

### 🐳 **Container Orchestration (ECS)**
- **What it is:** Manages multiple copies of your application running simultaneously
- **Why it helps:** Distributes load across multiple servers
- **Auto-scaling:** Automatically adds more servers when traffic increases

### ⚡ **Serverless Computing (Lambda)**
- **What it is:** Code that runs without managing servers
- **Benefits:** Pay only when processing orders, automatic scaling, zero maintenance
- **Trade-off:** Less control over the environment, but massive operational simplification

---

## How It Works

### 1. **Order Placement**
```bash
curl -X POST http://your-store.com/orders/async \
  -H "Content-Type: application/json" \
  -d '{"customer_id": 123, "items": [{"product_id": 1, "quantity": 2}]}'
```
**Result:** Instant response (< 100ms) with order confirmation

### 2. **Background Processing**
- Order gets published to SNS topic
- SQS queue receives the message
- Worker processes the order (3-second payment simulation)
- Customer receives confirmation

### 3. **Scaling Under Load**
- **Low traffic:** 1-2 servers handle everything
- **High traffic:** Auto-scaling adds more servers automatically
- **Peak traffic:** Lambda scales to thousands of concurrent executions

---

## Performance Results

### Synchronous Approach (Phase 1)
- **Orders processed:** 19 in 60 seconds
- **Customer wait time:** 29.5 seconds average
- **Success rate:** 100% of attempted orders (but many couldn't even be attempted)

### Asynchronous Approach (Phase 2)
- **Orders processed:** 3,500+ in 60 seconds
- **Customer wait time:** 33ms average
- **Success rate:** 100% of all orders
- **Improvement:** 184x more orders processed

### Serverless Approach (Phase 3)
- **Orders processed:** Same as Phase 2
- **Customer wait time:** 33ms average
- **Operational overhead:** Zero (vs manual scaling in Phase 2)
- **Cost:** FREE for startups under 267K orders/month

---

## Infrastructure Components

### 🌐 **Application Load Balancer (ALB)**
- **Purpose:** Distributes incoming requests across multiple servers
- **Health checks:** Ensures only healthy servers receive traffic
- **SSL termination:** Handles HTTPS encryption

### 🏗️ **ECS Fargate**
- **Purpose:** Runs containers without managing servers
- **Auto-scaling:** Automatically adjusts server count based on demand
- **Resource limits:** CPU and memory constraints per container

### 📨 **SNS (Simple Notification Service)**
- **Purpose:** Publishes order events to multiple subscribers
- **Reliability:** Guarantees message delivery
- **Fan-out:** One order can trigger multiple processes

### 📋 **SQS (Simple Queue Service)**
- **Purpose:** Stores orders until workers can process them
- **Durability:** Messages persist even if workers fail
- **Long polling:** Efficiently waits for new messages

### ⚡ **Lambda Functions**
- **Purpose:** Processes orders without managing servers
- **Scaling:** Automatically handles any load
- **Cost:** Pay only for actual processing time

---

## Testing Strategy

### 🧪 **Load Testing with Locust**
- **Tool:** Python-based load testing framework
- **Scenarios:** Normal load (5 users) vs Flash sale (20 users)
- **Metrics:** Response time, success rate, throughput

### 📊 **Monitoring with CloudWatch**
- **Metrics:** CPU usage, memory usage, queue depth
- **Alerts:** Notifications when systems approach limits
- **Dashboards:** Visual representation of system health

### 🔍 **Cold Start Analysis**
- **What:** Time to initialize Lambda function
- **Impact:** 188ms to 1409ms overhead on 3-second processing
- **Conclusion:** Negligible impact for payment processing

---

## Cost Analysis

### 💰 **ECS Approach**
- **Fixed cost:** $17/month (always running)
- **Scaling cost:** Additional servers as needed
- **Operational cost:** Manual monitoring and scaling

### 💰 **Lambda Approach**
- **Variable cost:** Pay per request
- **Free tier:** 1M requests + 400K GB-seconds monthly
- **Break-even:** ~1M orders/month vs ECS
- **Operational cost:** Zero (AWS manages everything)

### 📈 **Cost Comparison**
| Monthly Orders | ECS Cost | Lambda Cost | Savings
|----------------|----------|-------------|----------
| 10,000         | $17      | $0          | $17 (100%)
| 100,000        | $17      | $0          | $17 (100%)
| 267,000        | $17      | $0.08       | $16.92 (99.5%)
| 1,000,000      | $17      | $25.20      | -$8.20 (Lambda more expensive)

---

## Key Learnings

### 🎯 **When to Optimize vs Scale**
- **Optimize code:** When you can make algorithms faster
- **Scale infrastructure:** When you need more compute power
- **Use serverless:** When you want zero operational overhead

### 📈 **Scaling Strategies**
- **Vertical scaling:** Bigger servers (limited by hardware)
- **Horizontal scaling:** More servers (limited by coordination)
- **Serverless scaling:** Automatic scaling (limited by cost)

### 🔄 **Architecture Patterns**
- **Synchronous:** Simple but doesn't scale
- **Asynchronous:** Complex but scales well
- **Serverless:** Simple and scales automatically

---

## Getting Started

### Prerequisites
- AWS CLI configured
- Docker installed
- Terraform installed
- Go 1.23+ installed

### Quick Start
```bash
# 1. Deploy infrastructure
cd terraform
terraform init
terraform apply

# 2. Build and push containers
docker build -t your-ecr-repo/api-server .
docker push your-ecr-repo/api-server

# 3. Run load tests
cd test_locust
locust -f phase1_sync_test.py --host=http://your-alb-dns
```

### Testing Endpoints
- **Synchronous:** `POST /orders/sync` (waits for processing)
- **Asynchronous:** `POST /orders/async` (immediate response)
- **Health check:** `GET /health`

---

## Reports and Documentation

- **Homework 6 Report:** `Homework6.md` - Performance bottleneck analysis
- **Homework 7 Part 2 Report:** `HOMEWORK7_PART2_REPORT.txt` - Synchronous vs Asynchronous comparison
- **Homework 7 Part 3 Report:** `HW7_PART3_REPORT.txt` - Serverless Lambda analysis

---

## Team Contributions

**Hao Niu:** Part 3 implementation (Lambda serverless architecture)
- Lambda function development and deployment
- Cold start analysis and cost comparison
- Serverless architecture evaluation

**Aaron Wang:** Part 2 implementation (Asynchronous processing)
- ECS worker scaling experiments
- SNS/SQS integration
- Performance analysis and monitoring

---

## Conclusion

This project demonstrates the evolution from simple synchronous processing to sophisticated serverless architecture. We learned that:

1. **Synchronous systems fail under load** - customers wait too long
2. **Asynchronous systems scale well** - but require operational overhead
3. **Serverless systems eliminate complexity** - while maintaining performance

The key insight: **Modern cloud architecture isn't just about performance—it's about eliminating operational complexity while maintaining reliability and cost efficiency.**

For startups, serverless architecture provides massive cost savings and operational simplification, making it the clear choice for most use cases.