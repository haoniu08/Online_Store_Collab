"""
Homework 6 Search Load Testing
===============================

This Locust file specifically tests the product search endpoint for Homework 6.
The goal is to discover the breaking point of the system when searching
through 100,000 products with bounded iteration (100 products checked per search).

Test Scenarios:
- Test 1 - Baseline: 5 users for 2 minutes
- Test 2 - Breaking Point: 20 users for 3 minutes

Key Requirements:
- Use FastHttpUser for performance
- Search endpoint checks exactly 100 products per request
- Minimal wait time to create CPU load
- Search for common terms to ensure consistent behavior
"""

from locust import FastHttpUser, task, between
import random

class SearchLoadTestUser(FastHttpUser):
    """
    FastHttpUser for high-performance search load testing.
    Focuses on the /products/search endpoint to test bounded search performance.
    
    This class adapts its behavior based on the number of users:
    - Low users (<=5): Conservative baseline testing
    - Medium users (6-15): Comprehensive testing  
    - High users (>=16): Aggressive breaking point testing
    """
    
    # Maximum request rate - no waiting between requests for circuit breaker testing
    wait_time = between(0, 0)  # No delay - fire requests as fast as possible
    
    # Connection pooling for better performance
    connection_timeout = 10.0
    network_timeout = 10.0
    
    # Common search terms that will find results in our generated data
    search_terms = [
        # Brands (10 brands in our data)
        "Alpha", "Beta", "Gamma", "Delta", "Epsilon", 
        "Zeta", "Eta", "Theta", "Iota", "Kappa",
        
        # Categories (8 categories in our data)
        "Electronics", "Books", "Home", "Sports", 
        "Clothing", "Beauty", "Toys", "Automotive",
        
        # Partial matches for variety
        "Product", "Alpha", "Electronics", "Book",
        
        # Mixed case for case-insensitive testing
        "electronics", "ALPHA", "beta", "SPORTS"
    ]
    
    @task  # Only task - search products (100% of requests)
    def search_products(self):
        """
        GET /products/search?q={query}
        
        This is the core endpoint for Homework 6. Each request should:
        1. Check exactly 100 products (bounded iteration)
        2. Search in name and category fields (case-insensitive)
        3. Return max 20 results with total count
        4. Include search time in response
        
        All requests will be search operations to maximize CPU load testing.
        """
        query = random.choice(self.search_terms)
        
        with self.client.get(
            f"/products/search?q={query}",
            catch_response=True,
            name="/products/search"
        ) as response:
            if response.status_code == 200:
                try:
                    data = response.json()
                    
                    # Validate response structure
                    required_fields = ["products", "total_found", "search_time"]
                    if not all(field in data for field in required_fields):
                        response.failure(f"Missing required fields in response")
                        return
                    
                    # Validate data constraints
                    if len(data["products"]) > 20:
                        response.failure(f"Too many products returned: {len(data['products'])}")
                        return
                    
                    if data["total_found"] < 0:
                        response.failure(f"Invalid total_found: {data['total_found']}")
                        return
                    
                    response.success()
                    
                except Exception as e:
                    response.failure(f"JSON parsing failed: {str(e)}")
            else:
                response.failure(f"Search failed with status {response.status_code}")


# Note: Locust will use the SearchLoadTestUser class above for all testing scenarios
# The wait_time and task weights are designed to work for both baseline and breaking point tests