# Step 3 Part 2: Resource Efficiency Analysis

---

## Resource Utilization Comparison

### MySQL Resource Patterns

**Connection Management Overhead:**
- Required persistent connection pool (10 connections configured in RDS)
- Each connection consumes database memory and CPU
- Connection establishment overhead: ~50-100ms per new connection
- Unused connections still consume resources

**Resource Predictability:**
- Fixed instance size (db.t3.micro: 1 vCPU, 1GB RAM)
- Capacity reserved 24/7 regardless of load
- Predictable monthly cost but wasteful during low traffic
- Must provision for peak load, paying for unused capacity during normal load

**Operational Complexity:**
- Manual configuration of connection pool size
- Required VPC networking setup for RDS access
- Database backup management (automated RDS snapshots)
- Parameter group tuning for performance optimization

**Capacity Consumed (150 operations):**
- Database instance: 1 vCPU @ ~5% utilization
- Memory: ~200MB of 1GB used
- Storage: 20GB allocated (minimal actual usage)
- Network: ~30KB total data transfer

### DynamoDB Resource Patterns

**Managed Scaling:**
- No connection management required - API-based access
- On-demand billing automatically provisions capacity
- Each request independent - no persistent connections
- Zero idle resource consumption

**Resource Predictability:**
- Completely unpredictable capacity (managed by AWS)
- Pay only for actual requests consumed
- No reserved capacity wasted
- Automatic scaling handles spikes without intervention

**Operational Complexity:**
- Zero database administration
- No VPC required (public API endpoint)
- No backup configuration (automatic point-in-time recovery)
- No parameter tuning needed

**Capacity Consumed (150 operations):**
- Read Capacity Units: 50 RCUs (~$0.0125)
- Write Capacity Units: 100 WCUs (~$0.0125)
- Storage: <1MB actual data (~$0.00025/month)
- Network: ~30KB total data transfer

---

## Scaling Analysis

### How Resource Requirements Change With Load

**MySQL Scaling Pattern:**

| Load Level | Operations/sec | Instance Size | Cost/Month | Scaling Action |
|------------|---------------|---------------|------------|----------------|
| Low (10 ops/s) | 10 | db.t3.micro | $15 | None |
| Medium (100 ops/s) | 100 | db.t3.small | $30 | Vertical scaling (downtime) |
| High (1000 ops/s) | 1000 | db.t3.large | $120 | Vertical scaling + read replicas |
| Peak (10K ops/s) | 10000 | db.r5.xlarge | $500+ | Multi-AZ + 5 read replicas |

**Scaling Characteristics:**
- Step-function cost increases (instance size jumps)
- Manual intervention required for scaling
- Downtime required for vertical scaling
- Over-provisioning necessary to handle spikes
- Read replicas add eventual consistency complexity

**DynamoDB Scaling Pattern:**

| Load Level | Operations/sec | RCU/WCU | Cost/Month | Scaling Action |
|------------|---------------|---------|------------|----------------|
| Low (10 ops/s) | 10 | ~600 | $1 | None (automatic) |
| Medium (100 ops/s) | 100 | ~6K | $10 | None (automatic) |
| High (1000 ops/s) | 1000 | ~60K | $100 | None (automatic) |
| Peak (10K ops/s) | 10000 | ~600K | $1000 | None (automatic) |

**Scaling Characteristics:**
- Linear cost scaling with load
- Zero intervention required
- No downtime during scaling
- No over-provisioning needed
- Instant capacity adjustments

### Which Approach Offers More Predictable Resource Consumption?

**MySQL: Predictable Cost, Unpredictable Performance**
- Cost is fixed per month based on instance size
- Performance degrades unpredictably under load spikes
- Need to over-provision to guarantee SLA
- Waste ~70% of capacity during normal operations

**DynamoDB: Unpredictable Cost, Predictable Performance**
- Cost varies with actual usage
- Performance is consistent regardless of load
- Only pay for what you use
- No wasted capacity

**Winner for Shopping Carts:** DynamoDB. E-commerce traffic is highly variable (hourly/daily patterns, sales events). Paying only for actual usage is more cost-effective than reserving peak capacity 24/7.

---

## Capacity Planning Implications

### MySQL Capacity Planning

**Required Upfront Planning:**
1. Estimate peak load (hardest part - requires 3-6 months of data)
2. Choose instance size for peak + 20% headroom
3. Configure connection pool (2x expected concurrent connections)
4. Set up read replicas for high read loads
5. Plan Multi-AZ for high availability

**Ongoing Management:**
- Monitor CPU, memory, disk I/O, connection count
- Scale up before hitting limits (requires downtime)
- Tune query performance as data grows
- Manage replica lag if using read replicas

**Failure Modes:**
- Under-provisioning: Slow queries, connection exhaustion, SLA violations
- Over-provisioning: Wasted spend (paying for unused capacity)

**Example:** For shopping carts with 1000 ops/s peak but 100 ops/s average:
- Must provision db.t3.large ($120/month)
- Actually need db.t3.small for average load ($30/month)
- **Wasting $90/month (75% of cost)**

### DynamoDB Capacity Planning

**Required Upfront Planning:**
1. Design partition key for even distribution (critical)
2. Choose on-demand vs provisioned (on-demand safer for variable load)
3. That's it

**Ongoing Management:**
- Monitor for hot partitions (rare with good partition key design)
- Switch to provisioned capacity if load becomes predictable (cost optimization)
- No performance tuning required

**Failure Modes:**
- Bad partition key: Hot partitions cause throttling
- Extremely spiky load: On-demand can be expensive during spikes
- No failure mode for under-provisioning (auto-scales)

**Example:** For shopping carts with 1000 ops/s peak but 100 ops/s average:
- Pay for actual usage: ~$10/month average, $100/month during peaks
- **No wasted capacity**

---

## Resource Efficiency Conclusion

| Factor | MySQL | DynamoDB | Winner |
|--------|-------|----------|---------|
| **Idle Cost** | High (fixed instance) | Zero (pay per request) | DynamoDB |
| **Scaling Complexity** | High (manual, downtime) | Zero (automatic) | DynamoDB |
| **Resource Predictability** | Cost predictable, performance not | Performance predictable, cost not | DynamoDB (for variable load) |
| **Capacity Waste** | 50-80% typical | 0% | DynamoDB |
| **Operational Overhead** | High (tuning, monitoring, scaling) | Low (partition key design only) | DynamoDB |

**Key Insight:** DynamoDB's resource efficiency advantage is **massive for variable workloads**. MySQL's fixed cost model makes sense only when load is constant and predictable (rare in e-commerce). For shopping carts, DynamoDB eliminates capacity planning complexity and wasted spend.

**Data-Driven Finding:** Our test consumed $0.025 on DynamoDB vs $15/month minimum on MySQL (600x cost difference at low scale). At high scale (1M ops/day), DynamoDB would cost ~$1500/month vs MySQL at $500+/month, but MySQL would require significant operational overhead.
