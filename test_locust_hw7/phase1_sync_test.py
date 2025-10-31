"""
Phase 1: Synchronous Order Processing Load Tests

Test Configuration:
- Normal Operations: 5 concurrent users, 30 seconds
- Flash Sale: 20 concurrent users, 60 seconds
- User wait time: random 100-500ms between requests

Usage:
# Normal load test (5 users, 30s)
locust -f phase1_sync_test.py --host=http://localhost:8080 --users 5 --spawn-rate 1 --run-time 30s --headless

# Flash sale test (20 users, 60s)
locust -f phase1_sync_test.py --host=http://localhost:8080 --users 20 --spawn-rate 10 --run-time 60s --headless
"""

from locust import HttpUser, task, between
import random
import json


class OrderUser(HttpUser):
    """
    Simulates customers placing orders during normal operations and flash sales.
    Tests ONLY the synchronous endpoint (/orders/sync) for Phase 1.
    """

    # Wait time between requests: 100-500ms as specified in assignment
    wait_time = between(0.1, 0.5)

    @task
    def create_sync_order(self):
        """
        POST /orders/sync - Synchronous order processing
        Customer waits for payment verification (3 seconds)
        """
        # Generate random order with 1-3 items
        num_items = random.randint(1, 3)
        items = []

        for _ in range(num_items):
            items.append({
                "product_id": random.randint(1, 1000),
                "quantity": random.randint(1, 5),
                "price": round(random.uniform(9.99, 199.99), 2)
            })

        order_data = {
            "customer_id": random.randint(1, 10000),
            "items": items
        }

        with self.client.post(
            "/orders/sync",
            json=order_data,
            catch_response=True,
            name="POST /orders/sync"
        ) as response:
            if response.status_code == 200:
                try:
                    data = response.json()
                    if "order_id" in data and data["status"] == "completed":
                        response.success()
                    else:
                        response.failure("Invalid response structure")
                except json.JSONDecodeError:
                    response.failure("Invalid JSON response")
            else:
                response.failure(f"Got status code {response.status_code}")
