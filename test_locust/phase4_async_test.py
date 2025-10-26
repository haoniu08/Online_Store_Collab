"""
Phase 4: Asynchronous Order Processing - Queue Buildup Analysis

Test Configuration:
- Flash Sale: 20 concurrent users, 60 seconds
- User wait time: random 100-500ms between requests
- Endpoint: POST /orders/async

Expected Behavior:
- All requests accepted quickly (< 200ms)
- Queue builds up during test
- Monitor SQS ApproximateNumberOfMessagesVisible in CloudWatch

Usage:
# Flash sale test (20 users, 60s)
locust -f phase4_async_test.py --host=http://YOUR_ALB_DNS --users 20 --spawn-rate 10 --run-time 60s --headless
"""

from locust import HttpUser, task, between
import random
import json


class AsyncOrderUser(HttpUser):
    """
    Simulates customers placing orders using the asynchronous endpoint.
    This tests queue buildup during flash sale load.
    """

    # Wait time between requests: 100-500ms as specified in assignment
    wait_time = between(0.1, 0.5)

    @task
    def create_async_order(self):
        """
        POST /orders/async - Asynchronous order processing
        Customer gets immediate acknowledgment (< 100ms)
        Order is queued in SQS for background processing
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
            "/orders/async",
            json=order_data,
            catch_response=True,
            name="POST /orders/async"
        ) as response:
            # Async endpoint should return 202 Accepted
            if response.status_code == 202:
                try:
                    data = response.json()
                    if "order_id" in data and "message_id" in data:
                        response.success()
                    else:
                        response.failure("Missing order_id or message_id in response")
                except json.JSONDecodeError:
                    response.failure("Invalid JSON response")
            else:
                response.failure(f"Expected 202, got {response.status_code}")
