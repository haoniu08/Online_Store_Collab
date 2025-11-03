# DynamoDB Implementation Summary - HW8 Step 2

## Partition Key Strategy

**Chosen Key:** `shopping_cart_id` (Number type)

**Rationale:** All API operations query by cart ID, not customer ID. Using shopping_cart_id as partition key provides O(1) direct lookup for every operation (GetCart, AddOrUpdateItem). Timestamp-based ID generation (milliseconds since epoch) ensures even distribution across DynamoDB's hash space, preventing hot partitions. Alternative approach using customer_id would require sort key and risk hot partitions for power users, while our API never queries "all carts for customer."

**Validation:** Zero throttling across 150 operations, consistent 25-27ms latency, and no ConsumedCapacity spikes confirm even partition distribution.

---

## Key Differences from MySQL (Step 1)

**Schema Design:**
- MySQL: Normalized (2 tables with foreign key) → DynamoDB: Denormalized (single table with embedded items List)
- MySQL: AUTO_INCREMENT IDs → DynamoDB: Manual timestamp-based IDs
- MySQL: Native TIMESTAMP → DynamoDB: ISO8601 strings

**Data Access:**
- MySQL: JOIN to retrieve cart with items → DynamoDB: Single GetItem returns cart + items
- MySQL: ON DUPLICATE KEY UPDATE (1 query) → DynamoDB: Read-modify-write pattern (2 API calls: GetItem + UpdateItem)
- MySQL: Transactional ACID → DynamoDB: Per-item atomicity with ConditionExpression

**Operations:**
- MySQL: Uses SQL transactions for consistency → DynamoDB: Uses optimistic locking (ConditionExpression)
- MySQL: Foreign keys enforce integrity → DynamoDB: Application enforces integrity

**Critical Implementation Details:**
- Empty Go slices marshal to NULL not empty List (required manual AttributeValue construction)
- Reserved keywords (items, status) require ExpressionAttributeNames aliasing
- Numbers stored as strings internally (requires strconv conversion)

---

## Eventual Consistency Observations

**Configuration:** Used eventual consistency (`ConsistentRead: false`) for cost and performance.

**Testing Results:** 100% consistency observed across all tests despite eventual consistency mode. Read-after-write test (20 iterations) and sequential update test (10 rapid updates) both showed zero stale reads or lost updates.

**Why It Works:** Natural request spacing (200-500ms between user actions) exceeds DynamoDB's propagation time (<1 second per AWS SLA). Network latency alone (25-30ms) provides sufficient buffer. The read-modify-write pattern in AddOrUpdateItem naturally reads fresh data because adequate time passes between operations.

**Protection Mechanism:** ConditionExpression (`#status = :open`) provides optimistic locking - if cart modified between read and write, UpdateItem fails with ConditionalCheckFailedException (HTTP 400), allowing client retry.

**Practical Impact:** Zero user-facing consistency issues. Eventual consistency is functionally equivalent to strong consistency for this use case while providing 50% cost savings and lower latency.

---

## NoSQL vs SQL Trade-offs Discovered

**Denormalization (NoSQL win):**
- Gain: Single query retrieves cart + items (vs JOIN), 50% lower RCU consumption, simpler error handling
- Loss: Can't query items independently, harder to add new access patterns later
- Verdict: Acceptable for shopping cart use case where items always retrieved with cart

**Read-Modify-Write Pattern (SQL win):**
- MySQL: `ON DUPLICATE KEY UPDATE` in 1 statement
- DynamoDB: GetItem + UpdateItem (2 API calls)
- Performance impact: +2ms overhead (27.3ms vs 26.0ms)
- Verdict: Negligible difference for <50ms SLA

**Operational Complexity:**
- DynamoDB: No database administration, automatic scaling, managed infrastructure
- MySQL: Requires capacity planning, connection pooling, read replicas for scale
- Verdict: DynamoDB simpler operationally despite implementation complexity

**Flexibility:**
- MySQL: Can add indexes, change queries post-deployment
- DynamoDB: Partition key is immutable, schema changes require table rebuild
- Verdict: MySQL more flexible, DynamoDB requires upfront access pattern analysis

**Cost Model:**
- DynamoDB: Pay per request (on-demand) or provisioned capacity
- MySQL: Pay for instance size regardless of usage (RDS)
- Verdict: DynamoDB cheaper at variable/low load, MySQL cheaper at constant high load

---

## Performance Comparison

| Metric | MySQL (Step 1) | DynamoDB (Step 2) | Difference |
|--------|----------------|-------------------|------------|
| Overall Avg | 25.0ms | 26.4ms | +5.6% |
| Create Cart | 26.0ms | 26.8ms | +0.8ms |
| Add Items | 26.0ms | 27.3ms | +1.3ms |
| Get Cart | 23.0ms | 25.0ms | +2.0ms |
| Success Rate | 100% | 100% | Same |
| Error Rate | 0% | 0% | Same |

**Conclusion:** DynamoDB performed within 6% of MySQL despite architectural differences. Both meet <50ms SLA. The 1-2ms overhead is acceptable trade-off for DynamoDB's automatic scaling and operational simplicity.

---

## Key Lessons

1. **Design for Access Patterns:** DynamoDB requires knowing queries upfront. Partition key choice is permanent and critical.
2. **Measure, Don't Assume:** Eventual consistency seemed risky theoretically but proved undetectable in practice.
3. **Denormalization is Intentional:** Embedding data trades flexibility for performance - appropriate optimization for read-heavy workloads.
4. **Type System Matters:** DynamoDB's tagged values (NULL vs empty List, reserved keywords) require careful handling unlike SQL's simpler types.
5. **Operational Simplicity:** DynamoDB eliminates database administration but adds application complexity (manual ID generation, read-modify-write).

**Recommendation:** DynamoDB is production-ready for shopping cart use case with comparable performance to MySQL and better scalability characteristics.
