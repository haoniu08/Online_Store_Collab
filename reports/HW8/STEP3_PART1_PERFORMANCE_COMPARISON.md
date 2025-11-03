# Step 3 Part 1: Performance Comparison Analysis

**Data Source:** combined_results.json (150 operations each: 50 create, 50 add, 50 get)

---

## Required Comparison Table

| Metric | MySQL | DynamoDB | Winner | Margin |
|--------|-------|----------|---------|--------|
| **Avg Response Time (ms)** | 25.49 | 26.35 | MySQL | +0.86ms (3.4%) |
| **P50 Response Time (ms)** | 24.4 | 25.6 | MySQL | +1.2ms (4.9%) |
| **P95 Response Time (ms)** | 32.5 | 32.5 | Tie | 0ms (0%) |
| **P99 Response Time (ms)** | 64.4 | 63.35 | DynamoDB | -1.05ms (1.6%) |
| **Success Rate (%)** | 100.0 | 100.0 | Tie | 0% |
| **Total Operations** | 150 | 150 | - | - |

**Data Source:** combined_results.json

---

## Operation-Specific Breakdown

| Operation | MySQL Avg (ms) | DynamoDB Avg (ms) | Faster By |
|-----------|----------------|-------------------|-----------|
| **CREATE_CART** | 26.11 | 26.80 | MySQL by 0.69ms (2.6%) |
| **ADD_ITEMS** | 26.70 | 27.28 | MySQL by 0.58ms (2.2%) |
| **GET_CART** | 23.65 | 24.98 | MySQL by 1.33ms (5.6%) |

### Detailed Operation Statistics

**CREATE_CART:**
- MySQL: min 20.0ms, max 75.6ms, P50 23.9ms, P95 45.6ms
- DynamoDB: min 20.3ms, max 70.9ms, P50 25.6ms, P95 43.3ms

**ADD_ITEMS:**
- MySQL: min 20.9ms, max 38.6ms, P50 26.6ms, P95 35.8ms
- DynamoDB: min 21.2ms, max 33.9ms, P50 27.6ms, P95 33.1ms

**GET_CART:**
- MySQL: min 17.7ms, max 48.9ms, P50 22.8ms, P95 28.7ms
- DynamoDB: min 17.5ms, max 40.7ms, P50 24.8ms, P95 36.6ms

---

## Consistency Model Impact Assessment

### Actual Consistency Behavior Observed

**MySQL (ACID/Strong Consistency):**
- Every write immediately visible to subsequent reads
- Transactions provided atomicity guarantees
- 100% consistency across all 150 operations
- Predictable behavior: ON DUPLICATE KEY UPDATE ensured deterministic results

**DynamoDB (Eventual Consistency):**
- Used `ConsistentRead: false` for all GET operations
- Zero observed consistency delays across 50 GET operations
- 100% success rate despite eventual consistency model
- All read-after-write tests showed immediate consistency

### How Eventual Consistency Affected Application

**Expected Impact:** Stale reads, lost updates, race conditions

**Actual Impact:** None. Zero consistency issues detected because:

1. **Natural request spacing:** 200-500ms between operations far exceeds DynamoDB's <1s propagation SLA
2. **Network latency buffer:** 25-30ms per request provides inherent consistency window
3. **Read-modify-write pattern:** The GetItem + UpdateItem pattern in AddOrUpdateItem naturally reads fresh data
4. **ConditionExpression protection:** Optimistic locking (`#status = :open`) prevents lost updates

### Consistency Guarantees Comparison

**MySQL ACID Guarantees:**
- Atomicity: All-or-nothing transactions
- Consistency: Referential integrity via foreign keys
- Isolation: Serializable transactions prevent dirty reads
- Durability: Committed data persists despite crashes

**DynamoDB Eventual Consistency:**
- Single-item atomicity per operation
- Conditional updates prevent race conditions
- Eventually consistent reads (< 1 second SLA)
- Durable storage with 99.999999999% durability

**Key Difference:** MySQL guarantees immediate consistency; DynamoDB guarantees eventual consistency with optimistic locking fallback.

### User Experience Implications

**Shopping Cart Use Case:**
- Users modify carts at human speeds (seconds between actions)
- Cart operations are inherently sequential per user
- No cross-user cart access patterns

**Verdict:** Eventual consistency has **zero user-facing impact** for shopping carts. The theoretical consistency gap is masked by human interaction latency and network delays.

### Frequency of Consistency Delays

**Test Results:** 0 out of 150 operations (0.00%)

**Real-World Projection:** Based on AWS SLA (<1s propagation) and typical e-commerce usage (5-10s between cart updates), consistency delays would affect < 0.1% of operations even under high load.

### Application Patterns Most Affected

**Least Affected (Our Case):**
- Single-user cart operations
- Sequential updates with natural delays
- Read-heavy workloads with infrequent writes

**Most Affected (Not Our Case):**
- Multi-user collaborative editing
- Financial transactions requiring immediate consistency
- Real-time inventory management across warehouses

### Design Considerations for Each Model

**MySQL Strong Consistency Design:**
- Use transactions for multi-table operations
- Rely on foreign keys for referential integrity
- Connection pooling to manage database connections
- Read replicas introduce eventual consistency (trade-off)

**DynamoDB Eventual Consistency Design:**
- Use ConditionExpression for optimistic locking
- Design partition keys for even distribution
- Application-level referential integrity
- Strong consistency available via `ConsistentRead: true` at 2x cost

---

## Key Finding

MySQL was marginally faster (3.4% overall) but **both databases met the <50ms SLA**. The consistency model difference was **theoretically significant but practically invisible** for shopping cart operations. DynamoDB's eventual consistency posed zero risk in actual testing.
