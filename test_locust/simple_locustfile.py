from locust import HttpUser, FastHttpUser, task, between
import random
import json

class ProductAPIUser(HttpUser):
    """
    Simulates user behavior for the Product API.
    This will test both GET and POST endpoints with realistic patterns.
    """
    
    # Wait time between tasks (simulates user think time)
    wait_time = between(1, 3)  # 1-3 seconds between requests
    
    # Store product IDs we've created for GET requests
    product_ids = []
    
    def on_start(self):
        """
        Called when a simulated user starts.
        Pre-populate some products for testing.
        """
        # Create a few initial products
        for i in range(5):
            product_id = random.randint(1000, 9999)
            self.create_product(product_id)
            self.product_ids.append(product_id)
    
    def create_product(self, product_id):
        """Helper method to create a product"""
        product_data = {
            "product_id": product_id,
            "sku": f"SKU-{product_id}-{random.randint(100, 999)}",
            "manufacturer": random.choice([
                "Acme Corporation",
                "TechGear Inc",
                "Global Manufacturing",
                "Premium Products Ltd",
                "Innovation Industries"
            ]),
            "category_id": random.randint(1, 100),
            "weight": random.randint(100, 5000),
            "some_other_id": random.randint(1, 1000)
        }
        
        with self.client.post(
            f"/products/{product_id}/details",
            json=product_data,
            catch_response=True,
            name="/products/[id]/details"
        ) as response:
            if response.status_code == 204:
                response.success()
            else:
                response.failure(f"Got status code {response.status_code}")
    
    @task(10)  # Weight: 10 (most common operation)
    def get_product(self):
        """
        GET /products/{productId}
        This simulates browsing products - the most common operation in e-commerce.
        """
        if not self.product_ids:
            return
        
        product_id = random.choice(self.product_ids)
        
        with self.client.get(
            f"/products/{product_id}",
            catch_response=True,
            name="/products/[id]"
        ) as response:
            if response.status_code == 200:
                response.success()
                # Optionally validate response structure
                try:
                    data = response.json()
                    if "product_id" not in data:
                        response.failure("Missing product_id in response")
                except json.JSONDecodeError:
                    response.failure("Invalid JSON response")
            elif response.status_code == 404:
                response.failure("Product not found")
            else:
                response.failure(f"Got status code {response.status_code}")
    
    @task(3)  # Weight: 3 (less common than GET)
    def add_product(self):
        """
        POST /products/{productId}/details
        This simulates adding/updating products - less frequent than browsing.
        """
        product_id = random.randint(1, 100000)
        
        product_data = {
            "product_id": product_id,
            "sku": f"SKU-{product_id}-{random.randint(100, 999)}",
            "manufacturer": random.choice([
                "Acme Corporation",
                "TechGear Inc",
                "Global Manufacturing",
                "Premium Products Ltd",
                "Innovation Industries"
            ]),
            "category_id": random.randint(1, 100),
            "weight": random.randint(100, 5000),
            "some_other_id": random.randint(1, 1000)
        }
        
        with self.client.post(
            f"/products/{product_id}/details",
            json=product_data,
            catch_response=True,
            name="/products/[id]/details"
        ) as response:
            if response.status_code == 204:
                response.success()
                # Add to our list for future GET requests
                if product_id not in self.product_ids:
                    self.product_ids.append(product_id)
            else:
                response.failure(f"Got status code {response.status_code}")
    
    @task(1)  # Weight: 1 (rare)
    def health_check(self):
        """
        Check the health endpoint (used by load balancers).
        """
        with self.client.get("/health", name="/health", catch_response=True) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Health check failed: {response.status_code}")


class FastProductAPIUser(FastHttpUser):
    """
    FastHttpUser equivalent for performance comparison.
    Uses HTTP/1.1 keep-alive connections for better performance.
    """
    
    # Wait time between tasks (simulates user think time)
    wait_time = between(1, 3)  # 1-3 seconds between requests
    
    # Store product IDs we've created for GET requests
    product_ids = []
    
    # Connection settings for better performance
    connection_timeout = 10.0
    network_timeout = 10.0
    
    def on_start(self):
        """
        Called when a simulated user starts.
        Pre-populate some products for testing.
        """
        # Create a few initial products
        for i in range(5):
            product_id = random.randint(1000, 9999)
            self.create_product(product_id)
            self.product_ids.append(product_id)
    
    def create_product(self, product_id):
        """Helper method to create a product"""
        product_data = {
            "product_id": product_id,
            "sku": f"SKU-{product_id}-{random.randint(100, 999)}",
            "manufacturer": random.choice([
                "Acme Corporation",
                "TechGear Inc", 
                "Global Manufacturing",
                "Premium Products Ltd",
                "Innovation Industries"
            ]),
            "category_id": random.randint(1, 100),
            "weight": random.randint(100, 5000),
            "some_other_id": random.randint(1, 1000)
        }
        
        # FastHttpUser doesn't use context managers - simpler calls
        response = self.client.post(
            f"/products/{product_id}/details",
            json=product_data,
            name="/products/[id]/details"
        )
    
    @task(10)  # Weight: 10 (most common operation)
    def get_product(self):
        """
        GET /products/{productId}
        This simulates browsing products - the most common operation in e-commerce.
        """
        if not self.product_ids:
            return
        
        product_id = random.choice(self.product_ids)
        
        # FastHttpUser - simpler without context managers
        response = self.client.get(
            f"/products/{product_id}",
            name="/products/[id]"
        )
    
    @task(3)  # Weight: 3 (less common than GET)
    def add_product(self):
        """
        POST /products/{productId}/details
        This simulates adding/updating products - less frequent than browsing.
        """
        product_id = random.randint(1, 100000)
        
        product_data = {
            "product_id": product_id,
            "sku": f"SKU-{product_id}-{random.randint(100, 999)}",
            "manufacturer": random.choice([
                "Acme Corporation",
                "TechGear Inc",
                "Global Manufacturing", 
                "Premium Products Ltd",
                "Innovation Industries"
            ]),
            "category_id": random.randint(1, 100),
            "weight": random.randint(100, 5000),
            "some_other_id": random.randint(1, 1000)
        }
        
        # FastHttpUser - simpler without context managers
        response = self.client.post(
            f"/products/{product_id}/details",
            json=product_data,
            name="/products/[id]/details"
        )
        
        # Add to our list for future GET requests
        if product_id not in self.product_ids:
            self.product_ids.append(product_id)
    
    @task(1)  # Weight: 1 (rare)
    def health_check(self):
        """
        Check the health endpoint (used by load balancers).
        """
        response = self.client.get("/health", name="/health")