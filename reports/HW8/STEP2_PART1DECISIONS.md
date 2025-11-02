# DynamoDB Design Documentation - HW8 Step 2

## Design Overview

**Single-table design with embedded items**
- Partition Key: `shopping_cart_id` (Number)
- No sort key
- Items stored as embedded List attribute
- On-demand billing, eventual consistency
- No secondary indexes

**Performance: 27.7ms average vs MySQL 25.0ms (10% difference)**

---

## 1. Partition Key: shopping_cart_id

**Why this key?**
- All API operations query by cart ID (GetCart, AddOrUpdateItem)
- Ensures even distribution (timestamp-based IDs spread across partitions)
- Avoids hot partitions (vs customer_id which could concentrate traffic)
- Direct O(1) lookup for all operations

**Rejected: customer_id**
- Would require sort key to distinguish multiple carts per customer
- API doesn't query "all carts for customer"
- Risk of hot partitions for power users

---

## 2. No Sort Key

**Why not needed?**
- Each cart_id maps to exactly one cart (1:1 relationship)
- No range queries or multi-item lookups within partition
- Simpler = faster (26ms average GET latency)

---

## 3. Single Table with Embedded Items

**Structure:**
```
shopping_cart_id (Number, PK)
customer_id (Number)
status (String)
items (List of {product_id: Number, quantity: Number})
created_at (String)
updated_at (String)
```

**Why single table?**
- Always retrieve cart + items together (matches API behavior)
- One query vs two (lower latency, lower cost)
- Atomic updates via single UpdateItem call

**Rejected: Two-table design (cart + items separate)**
- Would need two queries or BatchGetItem
- 2x network latency, 2x cost
- More complexity, more failure modes

---

## 4. Cart-Item Relationship: Embedded List

**Implementation:**
```go
Items []cartItem `dynamodbav:"items"`
// Marshals to: {"L": [{"M": {"product_id": {"N": "123"}, "quantity": {"N": "2"}}}]}
```

**Why embedded List of Maps?**
- Retrieve all items in single GetItem call
- Update entire list atomically with UpdateItem
- Typical cart has <10 items (well under 400KB item limit)

**Trade-off accepted:**
- Can't query individual items independently
- Must read-modify-write entire list for updates
- Result: 2ms slower than MySQL (28ms vs 26ms for add_items)

---

## 5. No Secondary Indexes

**Why none needed?**
- API only queries by cart_id (never by customer_id, status, etc.)
- No range queries or filtering
- Indexes would double write cost with no benefit

**If we needed customer queries:**
Would add GSI on customer_id, but current API doesn't require it.

---

## 6. MySQL Comparison

| Aspect | MySQL (Step 1) | DynamoDB (Step 2) |
|--------|---------------|-------------------|
| Tables | 2 (normalized) | 1 (denormalized) |
| Items | Separate table, JOIN | Embedded List |
| ID generation | AUTO_INCREMENT | Timestamp-based |
| Updates | ON DUPLICATE KEY UPDATE | Read-modify-write |
| Consistency | Strong (transaction) | Eventual + optimistic locking |
| Latency | 25.0ms avg | 27.7ms avg |

**Key difference:** DynamoDB denormalizes for performance; MySQL normalizes for flexibility.

---

## 7. Key Trade-offs

**Denormalization for speed**
- Gain: 1 query instead of JOIN
- Loss: Can't query items independently

**Eventual consistency**
- Gain: 50% lower cost, lower latency
- Loss: Theoretical stale reads (never observed in practice)

**Read-modify-write pattern**
- Gain: Full control over item updates
- Loss: 2 API calls vs 1 SQL statement (2ms overhead)

---

## 8. Learning Journey

**Challenge 1: Empty items bug**
- Go empty slice marshals to NULL, not empty List
- Fixed: Manual construction with `{L: []*dynamodb.AttributeValue{}}`

**Challenge 2: Reserved keywords**
- "items" and "status" are reserved in DynamoDB
- Fixed: Use ExpressionAttributeNames (`#items`, `#status`)

**Challenge 3: Understanding optimistic locking**
- ConditionExpression prevents lost updates
- DynamoDB's alternative to transaction isolation

**Validation:**
- Smoke tests caught NULL bug early
- 150-operation test: 100% success, 27.7ms average
- 30-operation consistency test: 100% immediate visibility
- DynamoDB console inspection confirmed schema correctness

---

## Conclusion

**Design validates requirements:**
- <50ms latency achieved (27.7ms)
- Even partition distribution (no throttling)
- 100% consistency despite eventual consistency model
- Simple, cost-effective schema matching access patterns

**Production-ready for shopping cart use case at scale.**
