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
