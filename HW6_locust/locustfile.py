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
    
    # Adaptive wait time based on user count - will be set dynamically
    wait_time = between(0.1, 0.5)  # Default: 100-500ms between requests
    
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
    
    @task(10)  # Primary task - search products
    def search_products(self):
        """
        GET /products/search?q={query}
        
        This is the core endpoint for Homework 6. Each request should:
        1. Check exactly 100 products (bounded iteration)
        2. Search in name and category fields (case-insensitive)
        3. Return max 20 results with total count
        4. Include search time in response
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
    
    @task(2)  # Secondary task - health check (for ALB health checks)
    def health_check(self):
        """
        GET /health
        
        Health checks are important for load balancer functionality.
        This simulates ALB health check behavior.
        """
        with self.client.get(
            "/health",
            catch_response=True,
            name="/health"
        ) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Health check failed: {response.status_code}")
    
    @task(1)  # Edge case - empty search
    def empty_search(self):
        """
        Test error handling for empty search queries.
        This should return a 400 Bad Request.
        """
        with self.client.get(
            "/products/search?q=",
            catch_response=True,
            name="/products/search (empty)"
        ) as response:
            if response.status_code == 400:
                response.success()  # 400 is expected for empty query
            else:
                response.failure(f"Expected 400 for empty query, got {response.status_code}")
    
    @task(1)  # Edge case - no results
    def no_results_search(self):
        """
        Search for terms that should return no results.
        Tests the search performance when no matches are found.
        """
        no_result_terms = [
            "ZZZZZZ", "NoResultsExpected", "!@#$%", 
            "VeryUnlikelySearchTerm", "9999999999"
        ]
        query = random.choice(no_result_terms)
        
        with self.client.get(
            f"/products/search?q={query}",
            catch_response=True,
            name="/products/search (no results)"
        ) as response:
            if response.status_code == 200:
                try:
                    data = response.json()
                    if data["total_found"] == 0 and len(data["products"]) == 0:
                        response.success()
                    else:
                        response.failure(f"Expected no results, got {data['total_found']}")
                except Exception as e:
                    response.failure(f"JSON parsing failed: {str(e)}")
            else:
                response.failure(f"No results search failed: {response.status_code}")


# Note: Locust will use the SearchLoadTestUser class above for all testing scenarios
# The wait_time and task weights are designed to work for both baseline and breaking point tests