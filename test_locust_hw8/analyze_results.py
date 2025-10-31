#!/usr/bin/env python3
"""
Analyze mysql_test_results.json and compute summary metrics
"""
import json
import sys
from collections import defaultdict
import statistics

def analyze_results(filename):
    """Parse results and compute metrics"""
    results = []
    
    # Read JSON lines
    with open(filename, 'r') as f:
        for line in f:
            line = line.strip()
            if line:
                results.append(json.loads(line))
    
    if not results:
        print(f"No results found in {filename}")
        return
    
    # Group by operation
    by_operation = defaultdict(list)
    successes = 0
    failures = 0
    
    for r in results:
        by_operation[r['operation']].append(r['response_time'])
        if r['success']:
            successes += 1
        else:
            failures += 1
    
    # Overall metrics
    all_times = [r['response_time'] for r in results]
    
    print("="*60)
    print("HW8 MySQL Performance Test Results")
    print("="*60)
    print(f"\nTotal Operations: {len(results)}")
    print(f"  Success: {successes}")
    print(f"  Failure: {failures}")
    print(f"  Success Rate: {(successes/len(results)*100):.1f}%")
    
    print(f"\nOperation Breakdown:")
    for op, times in sorted(by_operation.items()):
        print(f"  {op}: {len(times)} operations")
    
    print(f"\nOverall Response Times (ms):")
    print(f"  Average: {statistics.mean(all_times):.1f}")
    print(f"  Median (P50): {statistics.median(all_times):.1f}")
    print(f"  P95: {sorted(all_times)[int(len(all_times)*0.95)]:.1f}")
    print(f"  P99: {sorted(all_times)[int(len(all_times)*0.99)]:.1f}")
    print(f"  Min: {min(all_times):.1f}")
    print(f"  Max: {max(all_times):.1f}")
    
    print(f"\nBy Operation Response Times (ms):")
    for op, times in sorted(by_operation.items()):
        print(f"  {op}:")
        print(f"    Average: {statistics.mean(times):.1f}")
        print(f"    Median: {statistics.median(times):.1f}")
        print(f"    P95: {sorted(times)[int(len(times)*0.95)]:.1f}")
        print(f"    Min: {min(times):.1f}")
        print(f"    Max: {max(times):.1f}")
    
    print("="*60)

if __name__ == "__main__":
    filename = sys.argv[1] if len(sys.argv) > 1 else "mysql_test_results.json"
    try:
        analyze_results(filename)
    except FileNotFoundError:
        print(f"Error: {filename} not found")
        print("Run the Locust test first to generate results")
        sys.exit(1)
    except Exception as e:
        print(f"Error analyzing results: {e}")
        sys.exit(1)
