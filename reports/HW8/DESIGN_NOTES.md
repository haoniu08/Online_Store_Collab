# HW8 Design Notes

## Step I (Infrastructure)
- RDS MySQL 8.0 on db.t3.micro
- Private subnets; SG allows 3306 only from ECS SG; not publicly accessible
- Skip final snapshot; deletion protection disabled (assignment)

## Part II (Schema & Data Layer)

### Schema Design
- Tables: `shopping_carts` (`shopping_cart_id` PK, `customer_id`, `status`, `created_at`, `updated_at`); `shopping_cart_items` ((`shopping_cart_id`, `product_id`) PK, `quantity`, timestamps)
- Rationale: parent cart + items; composite PK enforces one row per product per cart

### Key Strategy
- Primary Keys: BIGINT AUTO_INCREMENT for carts; composite for items
- Foreign Keys: `shopping_cart_items.shopping_cart_id` -> `shopping_carts.shopping_cart_id` (ON DELETE CASCADE)
- Constraints: `CHECK (quantity > 0)`; cart `status` limited to `OPEN|CHECKED_OUT|CANCELLED`

### Index Strategy
- `shopping_carts.customer_id` (history queries)
- `shopping_cart_items.shopping_cart_id` (fast join/retrieval)
- `shopping_cart_items.product_id` (optional; reporting/validation)

### Transaction Design
- Add/Update Item: single transaction using `INSERT ... ON DUPLICATE KEY UPDATE` (upsert)
- Validate cart exists and `status=OPEN` before mutate
- Isolation: READ COMMITTED; row locks scoped to affected rows

### Trade-offs Considered
- Composite PK vs surrogate `item_id`: chose composite for simpler upsert and uniqueness
- Quantity semantics: replace-by-value for predictability; could switch to `+=` if required
- No FK to products to decouple from catalog service; validate `product_id` at API layer if needed

## Part III (Implementation)

### Connection Pooling Configuration
- `SetMaxOpenConns(50)`: supports ~100 concurrent sessions with headroom
- `SetMaxIdleConns(10)`: reduces overhead while maintaining warm connections
- `SetConnMaxLifetime(30m)`: prevents stale connections; safe for RDS maintenance windows
- DSN includes `parseTime=true` for proper timestamp handling

### Transaction Handling
- Add/Update Item: single transaction wraps cart validation + upsert to ensure atomicity
- Used prepared statements (`db.PrepareContext`) to prevent SQL injection
- Rollback on any error; commit only after all operations succeed

### Error Handling & HTTP Status Codes
- 400 Bad Request: invalid input (missing customer_id, invalid quantity, malformed cart ID)
- 404 Not Found: cart not found or cart status != OPEN
- 500 Internal Server Error: DB connection/query failures
- Validated inputs before DB calls; return structured JSON error responses

### SQL Injection Prevention
- All queries use parameterized statements (`?` placeholders)
- No string concatenation or `fmt.Sprintf` in SQL
- Input validation layer rejects non-numeric IDs early

### Implementation Journey
- **Initial Approach**: Started with local MySQL (docker), applied migration, implemented repository layer, then wired handlers
- **Iterations**: Adjusted upsert semantics (replace vs increment quantity); settled on replace for predictability
- **Pool Tuning**: Initially used default pool settings; after local load testing adjusted MaxOpenConns to 50 based on connection count observations

### Performance Observations
- GET cart with JOIN: ~5-15ms locally (well under 50ms target)
- INSERT cart: ~2-5ms; upsert item: ~3-8ms
- No slow queries observed in initial testing; composite PK and indexes performed as expected

### Schema Modifications
- None required post-migration; initial design met all functional and performance requirements

### Valuable Learning Moments
- **Upsert Simplicity**: `INSERT ... ON DUPLICATE KEY UPDATE` eliminated need for separate SELECT + UPDATE logic
- **Connection Lifetime**: Setting `ConnMaxLifetime` prevents issues with RDS connection recycling
- **Composite PK Trade-off**: Simplified upsert but requires both `shopping_cart_id` + `product_id` in queries; acceptable for this access pattern
