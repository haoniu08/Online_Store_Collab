# Learning Notes - HW8 Step 2 Part 5

## What Surprised Me

**The NULL vs Empty List Bug Was Critical**
I expected Go's type system to handle empty slices cleanly, but DynamoDB distinguishes between NULL and empty List. When `[]cartItem{}` marshaled to `{"NULL": true}` instead of `{"L": []}`, all add_items operations failed with 500 errors because unmarshal couldn't convert NULL back to a slice. This required manual AttributeValue construction, which felt unnatural coming from ORMs where empty collections "just work."

**Reserved Keywords Are Pervasive**
I was surprised that "items" and "status" are DynamoDB reserved words. ValidationException errors weren't immediately obvious from the AWS SDK error messages. Learning to always use ExpressionAttributeNames felt verbose compared to SQL's simpler syntax, but it's necessary for DynamoDB's expression language.

**Eventual Consistency Was Invisible**
Despite theoretical concerns about stale reads, I observed zero consistency delays across 180 total operations. The gap between theory (eventual consistency could cause issues) and practice (request spacing naturally exceeds propagation time) was eye-opening. Network latency alone provides sufficient buffer for DynamoDB's sub-second propagation.

## Did My Partition Key Strategy Work?

Yes, using `shopping_cart_id` as the partition key worked exactly as intended. Timestamp-based IDs (milliseconds since epoch) distributed carts evenly across partitions with no throttling observed in CloudWatch metrics. Each cart ID maps to one cart, providing O(1) direct lookups. I initially considered `customer_id` but realized it would create hot partitions for power users and require a sort key to distinguish multiple carts. The final design matched our access pattern perfectly: all API operations query by cart_id, never by customer_id.

## NoSQL Concepts That Differed From Expectations

**Denormalization Is Correct, Not a Hack**
Coming from MySQL normalized schemas, embedding items as a List initially felt wrong. I expected to create separate `shopping_carts` and `cart_items` tables with foreign keys. Learning that DynamoDB is optimized for single-item retrieval with embedded data was a mindset shift. Denormalization reduces costs (1 RCU vs 2 RCUs) and latency (1 round-trip vs 2).

**Read-Modify-Write Instead of Native Upsert**
MySQL's `ON DUPLICATE KEY UPDATE` handles add-or-update in one SQL statement. DynamoDB required GetItem, modify in memory, then UpdateItem - two API calls. While this felt less elegant, ConditionExpression provides equivalent atomicity and the 2ms overhead is negligible.

**No Auto-Increment**
Generating cart IDs manually using timestamps felt primitive compared to MySQL's AUTO_INCREMENT. However, this forces better partition distribution (sequential IDs would create hot partitions) and works well at scale without coordination overhead.

## How Eventual Consistency Affected Testing

Eventual consistency didn't affect testing negatively at all. I designed consistency tests specifically to stress-test read-after-write scenarios with zero artificial delays. Despite using `ConsistentRead: false`, all 30 consistency tests passed immediately. This validated that natural request spacing (200-500ms) exceeds DynamoDB's propagation time (<1 second). The production test's 100% success rate across 150 operations confirmed eventual consistency is a non-issue for this use case. If anything, eventual consistency testing taught me to measure rather than assume - the theoretical concern didn't manifest in practice.

## Design Evolution

**Initial Approach: Normalized Two-Table Design**
I first designed separate tables: `shopping_carts` (PK: cart_id) and `cart_items` (PK: cart_id, SK: product_id). This felt natural coming from relational databases where normalization prevents redundancy. I expected to use BatchGetItem or Query to retrieve cart and items together.

**Why I Changed to Single-Table Design**
Performance testing revealed the two-table approach would require two DynamoDB queries or BatchGetItem complexity. Reading DynamoDB best practices, I learned that embedded data is preferred when items are always retrieved together. Since our API always returns cart with items (no "get cart without items" endpoint), embedding items as a List eliminated network overhead, halved RCU costs, and simplified error handling. The "waste" of storing items redundantly doesn't apply - there's only one copy per cart.

**Partition Key Evolution**
I briefly considered `customer_id` as partition key to enable "get all carts for customer" queries. However, our API spec has no such endpoint - we only query by cart_id. Using customer_id would require a sort key (cart_id) and complicate the most common operation (GetCart). I validated that shopping_cart_id alone provides even distribution by checking CloudWatch for throttling (none observed) and confirming timestamp-based IDs spread across numeric ranges.

## Hot Partition Issues

No hot partition issues encountered. I validated this through: (1) CloudWatch metrics showing zero ThrottledRequests across 150-operation test, (2) Even distribution of cart IDs across numeric ranges (1762077655xxx), (3) On-demand billing preventing throttling anyway, (4) Single-user-per-cart access pattern preventing traffic concentration. If I had used customer_id as partition key, a customer with 100 carts creating heavy load would concentrate traffic. Shopping_cart_id ensures each cart is an independent partition.

## How I Validated Design Choices

**Smoke Tests:** Created cart → added item → retrieved cart. Caught NULL items bug immediately when add_items returned 500. Fixed and retested until end-to-end flow succeeded.

**DynamoDB Console Inspection:** Examined actual stored items to verify structure. Confirmed items stored as `{"L": [{"M": {...}}]}` not NULL. Checked attribute types (N for numbers, S for strings). This visual confirmation was invaluable for debugging.

**Performance Testing:** Ran 150 operations (50 create, 50 add, 50 get) measuring latency and success rate. Validated 27.7ms average (within 10% of MySQL's 25.0ms), confirming design doesn't introduce unacceptable overhead.

**Consistency Testing:** Designed specific tests for read-after-write (20 iterations) and rapid updates (10 operations). Proved eventual consistency works despite theoretical concerns.

**CloudWatch Metrics:** Monitored RCU/WCU consumption, throttling, and error rates. Confirmed on-demand billing works efficiently at our scale with no capacity issues.

**Comparison Against MySQL:** Explicit A/B test between Step 1 (MySQL) and Step 2 (DynamoDB) using identical API and test harness. The 2.7ms difference validated that DynamoDB's different primitives (read-modify-write vs SQL upsert) don't significantly impact real-world performance.

## Key Insights Gained

1. **Measure, Don't Assume:** Eventual consistency seemed risky theoretically but proved undetectable in practice. Testing revealed the truth.

2. **Design for Access Patterns:** DynamoDB forces you to think about how data is queried. Unlike SQL where you can add indexes later, partition key choice is permanent. This constraint actually improves design by forcing clarity upfront.

3. **Trade-offs Are Explicit:** DynamoDB makes trade-offs visible (denormalization for speed, eventual consistency for cost), whereas SQL hides them (JOIN overhead, index maintenance). Neither is better universally - it depends on use case.

4. **Error Messages Matter:** ValidationException for reserved keywords and NULL marshaling issues weren't obvious from error text alone. DynamoDB console and CloudWatch logs were essential for debugging.

5. **NoSQL != No Structure:** DynamoDB's flexibility doesn't mean schema-less chaos. Strong typing in Go structs plus careful marshaling strategy provides structure equivalent to SQL schemas.
