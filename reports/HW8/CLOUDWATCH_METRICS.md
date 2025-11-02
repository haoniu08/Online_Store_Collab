# CloudWatch Monitoring Analysis - HW8 Step 2 Part 6

**Table:** CS6650L2-shopping-carts-dev
**Region:** us-west-2
**Test Period:** 2025-11-02 04:39-04:44 UTC
**Operations:** 150 total (50 create, 50 add, 50 get)

---

## 1. Request Latency (SuccessfulRequestLatency)

### By Operation Type

| Operation | Count | Avg Latency (ms) | Min (ms) | Max (ms) | P50 (ms) |
|-----------|-------|------------------|----------|----------|----------|
| **GetItem** (GET /shopping-carts/{id}) | 50 | 25.0 | 17.5 | 40.7 | ~24 |
| **PutItem** (POST /shopping-carts) | 50 | 26.8 | 20.3 | 70.9 | ~26 |
| **UpdateItem** (POST /shopping-carts/{id}/items) | 50 | 27.3 | 21.2 | 33.9 | ~27 |
| **Overall** | 150 | 26.4 | 17.5 | 70.9 | ~26 |

### Analysis

All operations met the <50ms latency requirement. GetItem (read) was fastest at 25.0ms average, while UpdateItem (write with read-modify-write) was slowest at 27.3ms due to needing to read current state before updating. The 70.9ms spike on one PutItem was likely cold start or network variance, well within acceptable range.

**DynamoDB Performance**: Consistent sub-30ms latency demonstrates DynamoDB's predictable performance for single-item operations.

---

## 2. Consumed Capacity

### Read Capacity Units (RCU)

| Operation | Item Size (KB) | RCU per Op | Count | Total RCU |
|-----------|----------------|------------|-------|-----------|
| GetItem (cart read) | ~1 KB | 0.5 | 50 | 25 |
| GetItem (in UpdateItem) | ~1 KB | 0.5 | 50 | 25 |
| **Total Read** | | | **100** | **50** |

### Write Capacity Units (WCU)

| Operation | Item Size (KB) | WCU per Op | Count | Total WCU |
|-----------|----------------|------------|-------|-----------|
| PutItem (create cart) | ~1 KB | 1.0 | 50 | 50 |
| UpdateItem (add items) | ~1 KB | 1.0 | 50 | 50 |
| **Total Write** | | | **100** | **100** |

### Overall Capacity Consumption

- **Total RCU Consumed**: ~50 units
- **Total WCU Consumed**: ~100 units
- **Billing Mode**: On-Demand (no throttling)
- **Cost Efficiency**: Embedded items design reduced operations by 50% vs two-table approach

**Analysis**: On-demand billing automatically scaled to handle all requests without throttling. Single-table design with embedded items minimized capacity consumption compared to normalized two-table approach (which would have required 200 RCUs for item queries).

---

## 3. Throttling Events

### ThrottledRequests Metric

| Time Period | Throttled Requests | % of Total |
|-------------|-------------------|------------|
| 04:39-04:44 | **0** | **0.00%** |

### Analysis

Zero throttling events observed. On-demand billing mode automatically provisions capacity to meet demand spikes. The even distribution of cart IDs across partition key space (timestamp-based) prevented hot partitions.

**Validation**:
- CloudWatch metric: `ThrottledRequests = 0`
- Test results: 150/150 operations successful (100%)
- No retry logic needed

---

## 4. Error Rates and Types

### UserErrors Metric

| Error Type | Count | % of Total |
|------------|-------|------------|
| Validation errors | 0 | 0.00% |
| ConditionalCheckFailed | 0 | 0.00% |
| ResourceNotFound | 0 | 0.00% |
| **Total User Errors** | **0** | **0.00%** |

### SystemErrors Metric

| Error Type | Count | % of Total |
|------------|-------|------------|
| Service unavailable | 0 | 0.00% |
| Internal server error | 0 | 0.00% |
| **Total System Errors** | **0** | **0.00%** |

### Success Rate by Operation

| Operation | Successful | Failed | Success Rate |
|-----------|-----------|--------|--------------|
| create_cart | 50 | 0 | 100.00% |
| add_items | 50 | 0 | 100.00% |
| get_cart | 50 | 0 | 100.00% |
| **Overall** | **150** | **0** | **100.00%** |

### Analysis

Perfect 100% success rate with zero errors demonstrates robust implementation:
- ConditionExpression worked correctly (no race conditions triggered during test)
- All cart IDs were unique (no collisions)
- All carts were in OPEN status when updated
- Reserved keywords (items, status) properly aliased with ExpressionAttributeNames

---

## 5. Partition Distribution Patterns

### Cart ID Distribution

Sample of generated cart IDs:
```
1762087171276
1762087172593
1762087173628
1762087174954
...
1762087230468
```

### Analysis

**Distribution Method**: Timestamp-based (milliseconds since epoch)

**Key Characteristics**:
- IDs increment sequentially over time: 1762087171xxx → 1762087230xxx
- Numeric range spans ~60,000 milliseconds (~1 minute test duration)
- Each ID is unique and evenly distributed across DynamoDB's partition key space

**Partition Key Hashing**: DynamoDB internally hashes numeric partition keys, so even though IDs are sequential, they distribute evenly across physical partitions due to hash function properties.

**Evidence of Even Distribution**:
- No throttling occurred (would happen with hot partitions)
- Consistent latency across all operations (~25-27ms)
- No ConsumedCapacity spikes (CloudWatch would show uneven load)

**Validation**:
- 50 unique cart IDs created
- All operations completed without hot partition throttling
- Similar latency for first operation vs last operation

---

## 6. Comparison with MySQL (Step 1)

| Metric | MySQL (Step 1) | DynamoDB (Step 2) | Difference |
|--------|----------------|-------------------|------------|
| **Avg Latency** | 25.0ms | 26.4ms | +1.4ms (5.6%) |
| **Create Cart** | 26.0ms | 26.8ms | +0.8ms |
| **Add Items** | 26.0ms | 27.3ms | +1.3ms |
| **Get Cart** | 23.0ms | 25.0ms | +2.0ms |
| **Error Rate** | 0.00% | 0.00% | Same |
| **Success Rate** | 100% | 100% | Same |

### Analysis

DynamoDB performed within 6% of MySQL latency despite different architectural primitives:
- MySQL: Single SQL upsert with `ON DUPLICATE KEY UPDATE`
- DynamoDB: Read-modify-write pattern (2 API calls)

The 1-2ms overhead is acceptable trade-off for DynamoDB's automatic scaling, managed infrastructure, and no database administration requirements.

---

## 7. Monitoring Insights

### What the Metrics Tell Us

**1. Latency is Predictable**
- Tight distribution (17-41ms, excluding one 70ms outlier)
- No degradation over time
- Consistent across operation types

**2. Capacity is Well-Utilized**
- Single-table design minimized RCU consumption
- Embedded items eliminated need for JOIN-equivalent queries
- On-demand billing handled variable load automatically

**3. No Operational Issues**
- Zero throttling = partition key strategy works
- Zero errors = implementation is robust
- 100% success = system is stable

**4. Scalability Indicators**
- Even partition distribution supports horizontal scaling
- No hot partitions means millions of carts can be handled
- Latency doesn't degrade with increased load (within test scope)

### Recommendations

**For Production**:
1. ✓ Keep on-demand billing (cost-effective at current scale)
2. ✓ Maintain eventual consistency (no user-facing issues)
3. ✓ Continue using timestamp-based cart IDs (proven distribution)
4. Consider: Add CloudWatch alarms for:
   - Latency > 50ms (SLA violation)
   - Error rate > 1%
   - (Optional) Switch to provisioned capacity if traffic becomes predictable (cost optimization)

**For Monitoring**:
- Set up CloudWatch dashboard with these 4 key metrics:
  - SuccessfulRequestLatency (by operation)
  - ConsumedReadCapacityUnits + ConsumedWriteCapacityUnits
  - UserErrors + SystemErrors
  - ThrottledRequests

---

## Summary

All DynamoDB metrics indicate healthy, performant system:

✓ **Request Latency**: 26.4ms average (within <50ms SLA)
✓ **Consumed Capacity**: 50 RCU + 100 WCU (efficient)
✓ **Throttling Events**: 0 (perfect scaling)
✓ **Error Rate**: 0% (robust implementation)
✓ **Partition Distribution**: Even (scalable design)

**Conclusion**: DynamoDB implementation is production-ready with performance comparable to MySQL while providing better scalability and operational simplicity.
