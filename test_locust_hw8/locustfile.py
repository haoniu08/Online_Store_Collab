"""
Locust performance test for HW8 MySQL Shopping Cart API
Runs exactly 150 operations: 50 create, 50 add items, 50 get
Outputs mysql_test_results.json in required format
"""
import json
import time
from datetime import datetime, timezone
from locust import HttpUser, task, events, between
from locust.runners import MasterRunner, WorkerRunner

# Global counters and results storage
operation_counts = {
    "create_cart": 0,
    "add_items": 0,
    "get_cart": 0
}
test_results = []
cart_ids = []  # Store created cart IDs for subsequent operations

# Target counts
TARGET_CREATES = 50
TARGET_ADDS = 50
TARGET_GETS = 50
TOTAL_OPERATIONS = TARGET_CREATES + TARGET_ADDS + TARGET_GETS


def record_result(operation, response_time, success, status_code):
    """Record a test result in the required format"""
    result = {
        "operation": operation,
        "response_time": round(response_time, 1),
        "success": success,
        "status_code": status_code,
        "timestamp": datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")
    }
    test_results.append(result)


@events.test_start.add_listener
def on_test_start(environment, **kwargs):
    """Initialize test - only run on master or standalone"""
    if not isinstance(environment.runner, WorkerRunner):
        print(f"Starting HW8 MySQL performance test: {TOTAL_OPERATIONS} operations")
        print(f"Target: {TARGET_CREATES} creates, {TARGET_ADDS} adds, {TARGET_GETS} gets")


@events.test_stop.add_listener
def on_test_stop(environment, **kwargs):
    """Write results to JSON file - only run on master or standalone"""
    if not isinstance(environment.runner, WorkerRunner):
        output_file = "mysql_test_results.json"
        
        # Write results as JSON array (matches homework format: array of objects)
        with open(output_file, 'w') as f:
            json.dump(test_results, f, indent=2)
        
        # Also verify we have exactly 150 operations
        if len(test_results) != TOTAL_OPERATIONS:
            print(f"  ⚠ WARNING: Expected {TOTAL_OPERATIONS} operations, got {len(test_results)}")
        
        print(f"\n{'='*60}")
        print(f"Test completed. Results written to {output_file}")
        print(f"Total operations: {len(test_results)}")
        print(f"  Creates: {operation_counts['create_cart']}")
        print(f"  Adds: {operation_counts['add_items']}")
        print(f"  Gets: {operation_counts['get_cart']}")
        print(f"{'='*60}\n")


class ShoppingCartUser(HttpUser):
    """
    User that performs shopping cart operations
    Stops after exact operation counts are reached
    """
    wait_time = between(0.1, 0.5)  # Small delay between operations
    
    def on_start(self):
        """Setup - runs once per user"""
        self.my_cart_ids = []
    
    @task(50)  # Higher weight to prioritize creates first
    def create_cart(self):
        """POST /shopping-carts - create new cart"""
        if operation_counts["create_cart"] >= TARGET_CREATES:
            return
        
        customer_id = int(time.time() * 1000) % 1000000  # Unique customer ID
        start_time = time.time()
        
        try:
            with self.client.post(
                "/shopping-carts",
                json={"customer_id": customer_id},
                catch_response=True
            ) as response:
                elapsed_ms = (time.time() - start_time) * 1000
                
                if response.status_code == 201:
                    operation_counts["create_cart"] += 1
                    data = response.json()
                    cart_id = data.get("shopping_cart_id")
                    if cart_id:
                        cart_ids.append(cart_id)
                        self.my_cart_ids.append(cart_id)
                    
                    record_result("create_cart", elapsed_ms, True, 201)
                    response.success()
                else:
                    record_result("create_cart", elapsed_ms, False, response.status_code)
                    response.failure(f"Expected 201, got {response.status_code}")
        except Exception as e:
            elapsed_ms = (time.time() - start_time) * 1000
            record_result("create_cart", elapsed_ms, False, 0)
    
    @task(30)
    def add_items_to_cart(self):
        """POST /shopping-carts/{id}/items - add item to cart"""
        if operation_counts["add_items"] >= TARGET_ADDS:
            return
        
        # Wait until we have carts created
        if not cart_ids:
            return
        
        # Use a cart ID (prefer own carts, fall back to any cart)
        cart_id = self.my_cart_ids[-1] if self.my_cart_ids else cart_ids[-1]
        product_id = (int(time.time() * 1000) % 100000) + 1  # Random product 1-100000
        quantity = 1 + (int(time.time()) % 5)  # Quantity 1-5
        
        start_time = time.time()
        
        try:
            with self.client.post(
                f"/shopping-carts/{cart_id}/items",
                json={"product_id": product_id, "quantity": quantity},
                catch_response=True
            ) as response:
                elapsed_ms = (time.time() - start_time) * 1000
                
                if response.status_code in (200, 201, 204):
                    operation_counts["add_items"] += 1
                    record_result("add_items", elapsed_ms, True, response.status_code)
                    response.success()
                else:
                    record_result("add_items", elapsed_ms, False, response.status_code)
                    response.failure(f"Expected 200/201/204, got {response.status_code}")
        except Exception as e:
            elapsed_ms = (time.time() - start_time) * 1000
            record_result("add_items", elapsed_ms, False, 0)
    
    @task(20)
    def get_cart(self):
        """GET /shopping-carts/{id} - retrieve cart with items"""
        if operation_counts["get_cart"] >= TARGET_GETS:
            return
        
        # Wait until we have carts created
        if not cart_ids:
            return
        
        # Use a cart ID
        cart_id = self.my_cart_ids[-1] if self.my_cart_ids else cart_ids[-1]
        
        start_time = time.time()
        
        try:
            with self.client.get(
                f"/shopping-carts/{cart_id}",
                catch_response=True
            ) as response:
                elapsed_ms = (time.time() - start_time) * 1000
                
                if response.status_code == 200:
                    operation_counts["get_cart"] += 1
                    record_result("get_cart", elapsed_ms, True, 200)
                    response.success()
                else:
                    record_result("get_cart", elapsed_ms, False, response.status_code)
                    response.failure(f"Expected 200, got {response.status_code}")
        except Exception as e:
            elapsed_ms = (time.time() - start_time) * 1000
            record_result("get_cart", elapsed_ms, False, 0)
    
    def on_stop(self):
        """Cleanup - runs when user stops"""
        # Check if we've reached target counts
        total = operation_counts["create_cart"] + operation_counts["add_items"] + operation_counts["get_cart"]
        if total >= TOTAL_OPERATIONS:
            # Stop the test
            self.environment.runner.quit()
