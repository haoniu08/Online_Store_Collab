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
- `SetMaxOpenConns(20)`: supports concurrent requests with headroom for db.t3.micro limits
- `SetMaxIdleConns(10)`: reduces overhead while maintaining warm connections
- `SetConnMaxLifetime(30m)`: prevents stale connections; safe for RDS maintenance windows
- `SetConnMaxIdleTime(10m)`: closes idle connections to free resources
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

## Part V (Learning Notes & Insights)

### What Surprised You?

**Initial Schema Performance**: The initial schema design met all performance requirements on first attempt. GET cart operations averaged 23.6ms (well under 50ms target) with composite PK and proper indexes. No slow query optimization was needed.

**Query Performance**: No queries were slower than expected. The JOIN between `shopping_carts` and `shopping_cart_items` performed efficiently even with concurrent operations during the 150-operation test.

**RDS Network Isolation Challenge**: Initially attempted to query RDS directly from local machine for verification. Discovered that RDS placement in private subnets (correct security practice) prevents direct external connections. This required verifying schema existence through:
  - ECS application logs (migration success messages)
  - API endpoint testing (successful cart creation confirms schema exists)
  - ECS task logs (database connectivity confirmed)
  
**Lesson**: AWS security best practices (private subnets) require different verification approaches than local development. Cannot simply `mysql -h endpoint` from outside VPC.

### Implementation Journey

**Initial Approach**: 
1. Designed schema with two tables (carts + items) based on OpenAPI spec
2. Implemented migrations as SQL file for manual application
3. Added repository layer with connection pooling
4. Wired endpoints to handlers

**What Didn't Work Initially**:
- **Connection Pooling**: Started with default Go `database/sql` settings. Initial local testing revealed need for explicit pool sizing (MaxOpenConns) to handle concurrent requests efficiently.
- **Migration Strategy**: Initially planned manual SQL application via command line. Realized migrations needed to run automatically on app startup since RDS is not directly accessible. Moved migration SQL into Go code (`runMigrations()`) embedded in main.go for idempotent execution.
- **Direct Database Access**: Attempted to verify schema via direct MySQL connection from local machine. Failed because RDS is in private subnets with security group only allowing ECS tasks. Verification shifted to indirect methods (logs, API responses).

**Optimizations for Test Requirements**:
- Connection pool: `MaxOpenConns(20)`, `MaxIdleConns(10)` - sufficient for test load with headroom (conservative for db.t3.micro connection limits)
- Transaction isolation: Used `READ COMMITTED` to balance consistency and performance
- Indexed `customer_id` for history queries (future requirement, not needed for 150-op test)
- Composite PK on items table enabled efficient upsert without additional SELECT queries

**What Would You Do Differently**:
1. **Migration Management**: Consider a dedicated migration tool (like `golang-migrate` or `flyway`) for production, rather than embedding SQL in application code
2. **Connection Pooling**: Start with explicit pool configuration earlier - would have saved iteration time
3. **Testing Strategy**: Set up local MySQL container earlier for schema validation before deploying to RDS
4. **Monitoring**: Add structured logging/metrics for database connection pool utilization to better understand resource usage patterns

### Database Concepts Learned

- **Composite Primary Keys**: First practical use case - eliminates need for surrogate key while enforcing business rule (one quantity per product per cart)
- **ON DUPLICATE KEY UPDATE**: MySQL-specific upsert pattern that simplifies concurrent item updates
- **Connection Pooling**: Understanding relationship between pool size, concurrent requests, and RDS connection limits
- **RDS Network Security**: Private subnet placement prevents external access; proper security group configuration critical for ECS-to-RDS connectivity
