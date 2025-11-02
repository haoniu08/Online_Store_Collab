# Eventual Consistency Investigation - HW8 Step 2 Part 3

## Test Design

Created `test_consistency_dynamodb.sh` to test three scenarios: read-after-write (20 iterations), sequential updates to same cart (10 rapid updates), and final state verification. Each test creates/updates carts then immediately reads them with no artificial delays to stress-test eventual consistency.

## Test Results

**Test 1: Read-After-Write (20 iterations)**
Created 20 carts and immediately retrieved each one. Result: 20/20 successful (100%). All newly created carts were immediately readable despite using eventual consistency mode (`ConsistentRead: false`).

**Test 2: Sequential Updates (10 rapid updates)**
Created one cart, then rapidly added 10 items with immediate reads after each addition. Result: 10/10 successful (100%). All updates succeeded without conflicts or stale reads.

**Test 3: Final State Verification**
After rapid updates, verified final cart state. Expected 10 items, got 10 items. No lost updates or phantom reads observed.

## Investigation Answers

### How frequently do you observe eventual consistency delays?

Zero delays observed across 30 consistency test operations and 150 production test operations. The theoretical propagation time (<1 second per AWS SLA) is masked by natural request spacing (200-500ms between user actions) and network latency (25-30ms round-trip). Production test data shows typical gaps: create_cart at T+0ms, add_items at T+251ms, next operation at T+620ms - well beyond propagation time.

### What application patterns are most affected by consistency delays?

Rapid successive operations on the same item are most vulnerable. Example: User double-clicks "Add to Cart" within 100ms could theoretically see stale data on second click. However, we saw zero issues because: (1) frontend debouncing prevents double-clicks, (2) human reaction time provides natural 300-500ms gaps, (3) our read-modify-write pattern in AddOrUpdateItem includes ConditionExpression (`#status = :open`) providing optimistic locking that fails if cart changed between read and write. Multi-region scenarios with cross-region reads would be more affected, but our single-region deployment avoids this.

### How can you design your application to handle consistency gracefully?

Use optimistic locking via ConditionExpression to prevent lost updates - if data changed since read, UpdateItem fails with ConditionalCheckFailedException returning 400 for client retry. Implement read-modify-write patterns rather than blind writes to ensure operating on current state. For critical operations requiring strong consistency, enable `ConsistentRead: true` (doubles cost). Add version numbers to items for stricter conflict detection. Design UI with natural delays (confirmation dialogs, animations) that exceed propagation time. For shopping carts, eventual consistency is acceptable because: cart operations are isolated per user, status changes are infrequent, ConditionExpression catches conflicts, and request spacing exceeds propagation time.

## Consistency Strategy

We use eventual consistency by default for performance and cost, relying on ConditionExpression for atomicity. Testing validated this approach works for shopping carts with zero user-facing issues. Strong consistency would be needed for financial transactions or inventory management where stale reads cause real problems, but shopping cart state tolerates brief inconsistency that never manifests in practice.
