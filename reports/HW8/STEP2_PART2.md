# API Implementation Analysis - HW8 Step 2 Part 2

## Question 1: DynamoDB Attribute Value Format Impact

DynamoDB uses tagged attribute values rather than native types: numbers are stored as `{"N": "123"}`, strings as `{"S": "text"}`, and lists as `{"L": [...]}`. This requires Go structs to be marshaled to/from DynamoDB format using `dynamodbattribute.MarshalMap()` and `UnmarshalMap()`, adding ~1-2ms overhead per operation.

A critical bug we encountered: empty Go slices marshal to NULL (`{"NULL": true}`) rather than empty Lists, causing unmarshal failures when reading back. The fix required manual construction with `"items": {L: []*dynamodb.AttributeValue{}}` instead of automatic marshaling (code: `dynamodb_cart_repository.go:63`).

DynamoDB has 500+ reserved keywords including "items" and "status". Direct use causes ValidationException. We must use ExpressionAttributeNames aliasing: `UpdateExpression: "SET #items = :items"` with `ExpressionAttributeNames: {"#items": "items"}` (code: `dynamodb_cart_repository.go:183-186`).

Numbers are stored as strings internally, requiring conversion via `strconv.FormatInt(cartID, 10)` for writes and parsing for reads. DynamoDB has no native timestamp type, so we store ISO8601 strings using `time.RFC3339` format and parse on read with `time.Parse()`.

---

## Question 2: DynamoDB Operations for Each Endpoint

### POST /shopping-carts - Create Cart
We use PutItem to create a new cart with all attributes in a single write. The ConditionExpression `attribute_not_exists(shopping_cart_id)` ensures uniqueness by preventing overwrites. If a collision occurs (ConditionalCheckFailedException), we retry with a new timestamp-based ID. Code: `dynamodb_cart_repository.go:69-76`.

### GET /shopping-carts/{id} - Retrieve Cart
We use GetItem for direct O(1) partition key lookup, the fastest read operation in DynamoDB. We specify `ConsistentRead: false` for eventual consistency (cheaper and faster). Query would be for reading multiple items within a partition; we only need one specific cart. Scan reads the entire table and is prohibitively expensive when we have the partition key. If `result.Item == nil`, we return 404. Code: `dynamodb_cart_repository.go:84-96`.

### POST /shopping-carts/{id}/items - Add/Update Items
We use a read-modify-write pattern: GetItem to read current state, modify the items list in memory to add new product or update existing quantity, then UpdateItem with ConditionExpression `#status = :open` for optimistic locking. This ensures the cart hasn't been closed since we read it. PutItem would replace the entire cart and lose other attributes. Errors: GetCart failure returns 404, ConditionalCheckFailedException returns 400 (cart not OPEN). Code: `dynamodb_cart_repository.go:118-194`.

---

## Question 3: Handling Eventual Consistency

### Configuration and Strategy

We use eventual consistency (`ConsistentRead: false`) for cost and performance. DynamoDB's propagation time is <1 second, while typical user request spacing is 200-500ms. Test results show create_cart at T+0ms, add_items at T+251ms, next add_items at T+620ms. Network latency alone (25-30ms) provides buffer, so propagation completes before the next request arrives.

The read-modify-write pattern in AddOrUpdateItem naturally reads fresh data because sufficient time has passed since the last write. ConditionExpression (`#status = :open`) provides optimistic locking - if the cart status changed between read and write, UpdateItem fails with ConditionalCheckFailedException, returning 400 to the client for retry.

### Test Results

Consistency testing across 30 operations showed 100% success: 20/20 read-after-write tests passed, 10/10 sequential updates to the same cart succeeded, with no stale reads or lost updates observed. The eventual consistency model proved invisible to users in practice.

### When Strong Consistency Would Be Needed

Strong consistency (`ConsistentRead: true`, doubles cost) is required for financial transactions, inventory management preventing overselling, multi-region cross-region reads, or compliance audit trails. For shopping carts, eventual consistency is sufficient and cost-effective.

### Error Handling

ConditionalCheckFailedException maps to HTTP 400 (cart not modifiable). ResourceNotFoundException indicates missing table (500). ValidationException catches reserved keyword issues (500). We log errors with `log.Printf("ERROR: AddOrUpdateItem failed for cart %d: %v", cartID, err)` for debugging (code: `shopping_cart.go:67`).

---

## Summary

DynamoDB's attribute value format requires 1-2ms marshaling overhead and careful handling of NULL vs empty List distinction. Reserved keywords need ExpressionAttributeNames aliasing. We use PutItem for CreateCart, GetItem for retrieval (O(1) lookup), and read-modify-write (GetItem + UpdateItem) for AddOrUpdateItem with optimistic locking. Eventual consistency works well because request spacing (200-500ms) exceeds propagation time (<1s), achieving 27.7ms average latency with 100% success rate.
