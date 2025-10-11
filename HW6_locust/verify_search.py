#!/usr/bin/env python3
"""
Verification script to test the search endpoint behavior.
This will help us understand if the bounded iteration is working correctly.
"""

import requests
import time
import json

# Configuration
HOST = "http://35.95.84.53:8080"
SEARCH_ENDPOINT = f"{HOST}/products/search"

def test_single_search(query):
    """Test a single search and measure response time"""
    start_time = time.time()
    
    try:
        response = requests.get(f"{SEARCH_ENDPOINT}?q={query}", timeout=10)
        end_time = time.time()
        
        if response.status_code == 200:
            data = response.json()
            response_time = (end_time - start_time) * 1000  # Convert to ms
            
            print(f"Query: '{query}'")
            print(f"  Status: {response.status_code}")
            print(f"  Response time: {response_time:.2f}ms")
            print(f"  Server search time: {data.get('search_time', 'N/A')}")
            print(f"  Products returned: {len(data.get('products', []))}")
            print(f"  Total found: {data.get('total_found', 'N/A')}")
            print(f"  First product: {data['products'][0]['name'] if data.get('products') else 'None'}")
            print()
            
            return response_time, data
        else:
            print(f"Error: {response.status_code} - {response.text}")
            return None, None
            
    except Exception as e:
        print(f"Request failed: {e}")
        return None, None

def test_cpu_load():
    """Test multiple searches to see CPU behavior"""
    queries = ["Alpha", "Electronics", "Product", "Beta", "Books"]
    
    print("=== Single Search Tests ===")
    for query in queries:
        test_single_search(query)
    
    print("\n=== CPU Load Test (10 rapid searches) ===")
    start_time = time.time()
    
    for i in range(10):
        response = requests.get(f"{SEARCH_ENDPOINT}?q=Alpha", timeout=5)
        if response.status_code != 200:
            print(f"Search {i+1} failed: {response.status_code}")
    
    end_time = time.time()
    total_time = end_time - start_time
    print(f"10 searches completed in {total_time:.2f}s")
    print(f"Average time per search: {(total_time/10)*1000:.2f}ms")

def test_edge_cases():
    """Test edge cases"""
    print("\n=== Edge Case Tests ===")
    
    # Empty query (should return 400)
    try:
        response = requests.get(f"{SEARCH_ENDPOINT}?q=", timeout=5)
        print(f"Empty query: {response.status_code}")
    except Exception as e:
        print(f"Empty query failed: {e}")
    
    # No results query
    test_single_search("ZZZZZZ")
    
    # Common query that should find many results
    test_single_search("Product")

if __name__ == "__main__":
    print("Testing Product Search API")
    print(f"Target: {HOST}")
    print("=" * 50)
    
    # First, test if the server is reachable
    try:
        health_response = requests.get(f"{HOST}/health", timeout=5)
        print(f"Health check: {health_response.status_code}")
        print()
    except Exception as e:
        print(f"Server not reachable: {e}")
        exit(1)
    
    test_cpu_load()
    test_edge_cases()
    
    print("\n=== Analysis ===")
    print("If response times are consistently <20ms with single searches,")
    print("but CPU usage is low with 20 users, possible issues:")
    print("1. Search is not checking exactly 100 products")
    print("2. Products are too simple to cause CPU load")
    print("3. Server has more CPU capacity than expected")
    print("4. Load balancer is distributing load across multiple instances")