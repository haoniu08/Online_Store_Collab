# Part II: Performance Bottleneck Analysis Report
**CS 6650 - Homework 6: Identifying When to Scale vs Optimize**

## Objective & Setup
We deployed a Go-based product search service to identify its breaking point and determine whether performance issues require code optimization or additional compute resources. The service implements bounded iteration, checking exactly 100 products per search across a dataset of 100,000 products.

**Infrastructure Configuration:**
- **ECS Fargate**: 256 CPU units (0.25 vCPU), 512 MB memory, 1 instance
- **Application**: Go service with sync.Map storage, FastHttpUser load testing
- **Search Logic**: Fixed computation checking exactly 100 products per request

## Key Findings: Unexpected Resilience

### Load Testing Results
Our testing revealed significantly higher capacity than expected:

| User Count | CPU Utilization | Response Time | System Status |
|------------|----------------|---------------|---------------|
| 5 users    | ~15%           | ~20ms         | Excellent     |
| 20 users   | ~10%           | <20ms        | Good          |
| 100 users  | ~80%           | <30ms        | Acceptable    |
| 5000 users | >95%           | >2000ms       | **Breaking Point** |

### Critical Discovery
**The system did not break at 20 users as anticipated**, but remained stable until approximately **5000 concurrent users**. This indicates our implementation was far more efficient than the baseline expectations.

## Analysis: Why the System Performed Better

### 1. Efficient Implementation
- **Bounded iteration** working correctly (exactly 100 products checked per search)
- **Go's performance**: Excellent concurrency handling and string operations
- **Memory efficiency**: sync.Map providing thread-safe access without significant overhead
- **Simple search logic**: Case-insensitive string matching is highly optimized

### 2. Resource Utilization Pattern
- **CPU**: Linear scaling until ~5000 users, then exponential degradation
- **Memory**: Remained steady (~512MB) throughout all tests
- **Network**: Not a bottleneck until extreme loads

### 3. Evidence for Scaling vs Optimization Decision

**Indicators pointing to SCALING needed (not code optimization):**
- ✅ **Linear performance degradation**: Response times scaled predictably with load
- ✅ **CPU-bound bottleneck**: CPU hit 100% while memory remained stable  
- ✅ **Fixed computation per request**: Each search does exactly 100 product checks
- ✅ **No obvious algorithmic inefficiencies**: Search logic is simple and direct

**This suggests the solution is horizontal scaling (more servers) rather than code optimization.**

## CloudWatch Monitoring Evidence
- **CPU Utilization**: Clear correlation between user count and CPU usage
- **Memory Utilization**: Stable across all load levels (good memory management)
- **Response Time Metrics**: Exponential increase only at extreme loads (5000+ users)

## Stress Testing Variations
We conducted creative stress testing scenarios:
- **Burst testing**: Sudden spikes from 10 to 5000 users
- **Sustained load**: 1000+ users for extended periods
- **Search pattern variations**: Different query types and frequencies
- **Edge case testing**: Empty queries, no-result searches

## Conclusion: Scale vs Optimize Decision

**Decision: HORIZONTAL SCALING is needed**

**Evidence:**
1. **System performed 250x better than expected** (5000 vs 20 users)
2. **CPU-bound bottleneck** with efficient memory usage
3. **Fixed computational work per request** cannot be optimized further
4. **Linear scaling until saturation point** indicates well-designed system

**Recommendation:** Deploy multiple instances with load balancing rather than attempting code optimization. The current implementation is already highly efficient - the bottleneck is simply insufficient compute capacity for extreme loads.

---

# Part III: Horizontal Scaling Solution

## Implementation: ALB + Auto Scaling Architecture

Following our Part II analysis, we implemented horizontal scaling to solve the 5000+ user bottleneck:

**Infrastructure Components:**
- **Application Load Balancer (ALB)**: Distributes requests across multiple healthy instances
- **Target Group**: IP-based targeting for Fargate tasks with `/health` checks every 30s
- **Auto Scaling Policy**: CPU-based scaling (70% target, 2-4 instances, 300s cooldown)
- **Enhanced Security**: ALB security group (public HTTP) + ECS security group (ALB-only access)

## Horizontal Scaling Results

### Load Testing with ALB
Testing the same loads that caused issues in Part II:

| User Count | Single Instance | ALB + Auto Scaling | Improvement |
|------------|----------------|-------------------|-------------|
| 100 users  | <30ms, ~80% CPU | <25ms, ~40% CPU | Better distribution |
| 500 users  | Degraded performance | <50ms, stable | Significant improvement |
| 5000 users | >2000ms, system strain | <200ms, auto-scaled to 4 instances | **Problem Solved** |

### Auto Scaling Behavior Observed
- **Scale-out trigger**: CPU >70% sustained for >5 minutes
- **New instances**: Automatically added during load spikes
- **Load distribution**: Even spread across healthy targets
- **Scale-in**: Gradual reduction when load decreased

## Component Analysis

### 1. Application Load Balancer (ALB)
- **Role**: Traffic distribution and health monitoring
- **Benefit**: Single public endpoint, health-based routing
- **Observed**: Even load distribution, automatic unhealthy target removal

### 2. Auto Scaling Policy
- **Role**: Dynamic capacity adjustment based on CPU metrics
- **Configuration**: 70% CPU target, 2-4 instances, 5-minute cooldowns
- **Observed**: Responsive scaling during load tests, prevented resource waste

### 3. Target Group Health Checks
- **Role**: Ensures only healthy instances receive traffic
- **Configuration**: `/health` endpoint, 30s intervals, 2 consecutive successes
- **Observed**: Quick detection of unhealthy instances, seamless failover

## Resilience Testing Results

**Instance Failure Simulation:**
- Manually stopped 1 instance during 1000-user load test
- **Result**: Zero downtime, ALB immediately routed traffic to healthy instances
- **Recovery**: New instance auto-launched within 2 minutes

## Horizontal vs Vertical Scaling Trade-offs

| Aspect | Horizontal Scaling (Our Solution) | Vertical Scaling |
|--------|-----------------------------------|------------------|
| **Capacity** | Linear scaling with instances | Limited by hardware |
| **Resilience** | High (instance failures handled) | Single point of failure |
| **Cost** | Pay-per-use, efficient | Over-provisioning required |
| **Complexity** | Higher (ALB, auto-scaling) | Simpler setup |
| **Our Use Case** | ✅ Perfect fit - CPU-bound workload | ❌ Would hit limits |

## Key Success Metrics

**Problem Resolution:**
- ✅ **5000-user load**: Handled successfully with auto-scaling
- ✅ **Response times**: Maintained <200ms under extreme load  
- ✅ **System resilience**: Zero downtime during instance failures
- ✅ **Cost efficiency**: Scales down automatically during low load

**Evidence for Horizontal Scaling Success:**
1. **Solved the bottleneck**: System now handles loads that previously caused failures
2. **Maintained performance**: Response times remain acceptable under high load
3. **Added resilience**: Individual instance failures don't affect service availability
4. **Cost-effective**: Only pay for resources when needed

## Conclusion: Horizontal Scaling Validation

The horizontal scaling solution successfully addressed our Part II findings:

- **Original bottleneck**: Single instance CPU saturation at 5000+ users
- **Solution effectiveness**: Multi-instance architecture with load balancing
- **Performance improvement**: >90% response time improvement under heavy load
- **Operational benefits**: Automatic scaling, fault tolerance, cost optimization

This validates our Part II analysis - the system was well-optimized and needed more compute capacity, not code changes. Horizontal scaling was the correct architectural solution.

---
*This demonstrates the power of horizontal scaling for CPU-bound workloads - sometimes the best optimization is simply distributing the work across multiple instances.*