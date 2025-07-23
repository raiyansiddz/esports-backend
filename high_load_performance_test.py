#!/usr/bin/env python3
"""
HIGH-LOAD PERFORMANCE TESTING for GoLang eSports Fantasy Backend
Tests 1000 concurrent requests with detailed performance metrics
"""

import requests
import json
import sys
import os
import time
import threading
import statistics
import psutil
from datetime import datetime
from concurrent.futures import ThreadPoolExecutor, as_completed
from collections import defaultdict
import queue

# Backend configuration
BACKEND_URL = "http://localhost:8001"
API_BASE_URL = f"{BACKEND_URL}/api/v1"

# Performance test configuration
CONCURRENT_REQUESTS = 1000
THREAD_POOL_SIZE = 100
REQUEST_TIMEOUT = 30

print(f"🚀 HIGH-LOAD PERFORMANCE TESTING")
print(f"Backend URL: {BACKEND_URL}")
print(f"Concurrent Requests: {CONCURRENT_REQUESTS}")
print(f"Thread Pool Size: {THREAD_POOL_SIZE}")
print("=" * 80)

class PerformanceMetrics:
    def __init__(self):
        self.response_times = []
        self.status_codes = defaultdict(int)
        self.errors = []
        self.start_time = None
        self.end_time = None
        self.successful_requests = 0
        self.failed_requests = 0
        self.db_query_times = []
        self.memory_usage = []
        self.cpu_usage = []
        
    def add_response(self, response_time, status_code, error=None):
        self.response_times.append(response_time)
        self.status_codes[status_code] += 1
        if error:
            self.errors.append(error)
            self.failed_requests += 1
        else:
            self.successful_requests += 1
    
    def add_system_metrics(self, memory_percent, cpu_percent):
        self.memory_usage.append(memory_percent)
        self.cpu_usage.append(cpu_percent)
    
    def calculate_stats(self):
        if not self.response_times:
            return {}
        
        total_time = self.end_time - self.start_time if self.end_time and self.start_time else 0
        
        return {
            'total_requests': len(self.response_times),
            'successful_requests': self.successful_requests,
            'failed_requests': self.failed_requests,
            'success_rate': (self.successful_requests / len(self.response_times)) * 100,
            'total_duration': total_time,
            'requests_per_second': len(self.response_times) / total_time if total_time > 0 else 0,
            'avg_response_time': statistics.mean(self.response_times),
            'median_response_time': statistics.median(self.response_times),
            'min_response_time': min(self.response_times),
            'max_response_time': max(self.response_times),
            'p95_response_time': self.percentile(self.response_times, 95),
            'p99_response_time': self.percentile(self.response_times, 99),
            'status_codes': dict(self.status_codes),
            'error_rate': (self.failed_requests / len(self.response_times)) * 100,
            'avg_memory_usage': statistics.mean(self.memory_usage) if self.memory_usage else 0,
            'max_memory_usage': max(self.memory_usage) if self.memory_usage else 0,
            'avg_cpu_usage': statistics.mean(self.cpu_usage) if self.cpu_usage else 0,
            'max_cpu_usage': max(self.cpu_usage) if self.cpu_usage else 0,
        }
    
    @staticmethod
    def percentile(data, percentile):
        if not data:
            return 0
        sorted_data = sorted(data)
        index = int((percentile / 100) * len(sorted_data))
        return sorted_data[min(index, len(sorted_data) - 1)]

def monitor_system_resources(metrics, stop_event):
    """Monitor system resources during the test"""
    while not stop_event.is_set():
        try:
            memory_percent = psutil.virtual_memory().percent
            cpu_percent = psutil.cpu_percent(interval=0.1)
            metrics.add_system_metrics(memory_percent, cpu_percent)
            time.sleep(0.5)
        except Exception as e:
            print(f"Error monitoring system resources: {e}")
            break

def make_request(endpoint, method='GET', data=None, headers=None):
    """Make a single HTTP request and measure performance"""
    start_time = time.time()
    try:
        if method == 'GET':
            response = requests.get(endpoint, timeout=REQUEST_TIMEOUT, headers=headers)
        elif method == 'POST':
            response = requests.post(endpoint, json=data, timeout=REQUEST_TIMEOUT, headers=headers)
        
        end_time = time.time()
        response_time = end_time - start_time
        
        return {
            'response_time': response_time,
            'status_code': response.status_code,
            'error': None,
            'response_size': len(response.content) if response.content else 0
        }
    except Exception as e:
        end_time = time.time()
        response_time = end_time - start_time
        return {
            'response_time': response_time,
            'status_code': 0,
            'error': str(e),
            'response_size': 0
        }

def test_health_check_load(num_requests):
    """Test health check endpoint under load"""
    print(f"\n🏥 HEALTH CHECK LOAD TEST ({num_requests} requests)")
    print("-" * 60)
    
    metrics = PerformanceMetrics()
    stop_event = threading.Event()
    
    # Start system monitoring
    monitor_thread = threading.Thread(target=monitor_system_resources, args=(metrics, stop_event))
    monitor_thread.start()
    
    metrics.start_time = time.time()
    
    with ThreadPoolExecutor(max_workers=THREAD_POOL_SIZE) as executor:
        futures = []
        for i in range(num_requests):
            future = executor.submit(make_request, f"{BACKEND_URL}/health")
            futures.append(future)
        
        for future in as_completed(futures):
            result = future.result()
            metrics.add_response(
                result['response_time'],
                result['status_code'],
                result['error']
            )
    
    metrics.end_time = time.time()
    stop_event.set()
    monitor_thread.join()
    
    return metrics.calculate_stats()

def test_otp_generation_load(num_requests):
    """Test OTP generation endpoint under load (database writes)"""
    print(f"\n📱 OTP GENERATION LOAD TEST ({num_requests} requests)")
    print("-" * 60)
    
    metrics = PerformanceMetrics()
    stop_event = threading.Event()
    
    # Start system monitoring
    monitor_thread = threading.Thread(target=monitor_system_resources, args=(metrics, stop_event))
    monitor_thread.start()
    
    metrics.start_time = time.time()
    
    with ThreadPoolExecutor(max_workers=THREAD_POOL_SIZE) as executor:
        futures = []
        for i in range(num_requests):
            phone_number = f"+91999999{i:04d}"  # Generate unique phone numbers
            data = {"phone_number": phone_number}
            future = executor.submit(
                make_request, 
                f"{API_BASE_URL}/auth/send-otp", 
                'POST', 
                data,
                {"Content-Type": "application/json"}
            )
            futures.append(future)
        
        for future in as_completed(futures):
            result = future.result()
            metrics.add_response(
                result['response_time'],
                result['status_code'],
                result['error']
            )
    
    metrics.end_time = time.time()
    stop_event.set()
    monitor_thread.join()
    
    return metrics.calculate_stats()

def test_tournament_listing_load(num_requests):
    """Test tournament listing endpoint under load (database reads)"""
    print(f"\n🏆 TOURNAMENT LISTING LOAD TEST ({num_requests} requests)")
    print("-" * 60)
    
    metrics = PerformanceMetrics()
    stop_event = threading.Event()
    
    # Start system monitoring
    monitor_thread = threading.Thread(target=monitor_system_resources, args=(metrics, stop_event))
    monitor_thread.start()
    
    metrics.start_time = time.time()
    
    with ThreadPoolExecutor(max_workers=THREAD_POOL_SIZE) as executor:
        futures = []
        for i in range(num_requests):
            future = executor.submit(make_request, f"{API_BASE_URL}/admin/tournaments")
            futures.append(future)
        
        for future in as_completed(futures):
            result = future.result()
            metrics.add_response(
                result['response_time'],
                result['status_code'],
                result['error']
            )
    
    metrics.end_time = time.time()
    stop_event.set()
    monitor_thread.join()
    
    return metrics.calculate_stats()

def test_tournament_creation_load(num_requests):
    """Test tournament creation endpoint under load (complex database operations)"""
    print(f"\n🎮 TOURNAMENT CREATION LOAD TEST ({num_requests} requests)")
    print("-" * 60)
    
    metrics = PerformanceMetrics()
    stop_event = threading.Event()
    
    # Start system monitoring
    monitor_thread = threading.Thread(target=monitor_system_resources, args=(metrics, stop_event))
    monitor_thread.start()
    
    metrics.start_time = time.time()
    
    with ThreadPoolExecutor(max_workers=THREAD_POOL_SIZE) as executor:
        futures = []
        for i in range(num_requests):
            tournament_data = {
                "name": f"Load Test Tournament {i}",
                "game": "Valorant",
                "start_date": "2024-12-01T10:00:00Z",
                "end_date": "2024-12-15T18:00:00Z",
                "description": f"Load test tournament #{i}",
                "prize_pool": 100000.0 + i,
                "max_teams": 16
            }
            future = executor.submit(
                make_request, 
                f"{API_BASE_URL}/admin/tournaments", 
                'POST', 
                tournament_data,
                {"Content-Type": "application/json"}
            )
            futures.append(future)
        
        for future in as_completed(futures):
            result = future.result()
            metrics.add_response(
                result['response_time'],
                result['status_code'],
                result['error']
            )
    
    metrics.end_time = time.time()
    stop_event.set()
    monitor_thread.join()
    
    return metrics.calculate_stats()

def test_mixed_workload(num_requests):
    """Test mixed workload simulation"""
    print(f"\n🔄 MIXED WORKLOAD SIMULATION ({num_requests} requests)")
    print("-" * 60)
    
    metrics = PerformanceMetrics()
    stop_event = threading.Event()
    
    # Start system monitoring
    monitor_thread = threading.Thread(target=monitor_system_resources, args=(metrics, stop_event))
    monitor_thread.start()
    
    metrics.start_time = time.time()
    
    with ThreadPoolExecutor(max_workers=THREAD_POOL_SIZE) as executor:
        futures = []
        
        for i in range(num_requests):
            # Mix different types of requests
            request_type = i % 4
            
            if request_type == 0:  # Health check (40% of requests)
                future = executor.submit(make_request, f"{BACKEND_URL}/health")
            elif request_type == 1:  # OTP generation (30% of requests)
                phone_number = f"+91888888{i:04d}"
                data = {"phone_number": phone_number}
                future = executor.submit(
                    make_request, 
                    f"{API_BASE_URL}/auth/send-otp", 
                    'POST', 
                    data,
                    {"Content-Type": "application/json"}
                )
            elif request_type == 2:  # Tournament listing (20% of requests)
                future = executor.submit(make_request, f"{API_BASE_URL}/admin/tournaments")
            else:  # Tournament creation (10% of requests)
                tournament_data = {
                    "name": f"Mixed Load Tournament {i}",
                    "game": "CS:GO",
                    "start_date": "2024-12-01T10:00:00Z",
                    "end_date": "2024-12-15T18:00:00Z",
                    "description": f"Mixed load test tournament #{i}",
                    "prize_pool": 50000.0 + i,
                    "max_teams": 8
                }
                future = executor.submit(
                    make_request, 
                    f"{API_BASE_URL}/admin/tournaments", 
                    'POST', 
                    tournament_data,
                    {"Content-Type": "application/json"}
                )
            
            futures.append(future)
        
        for future in as_completed(futures):
            result = future.result()
            metrics.add_response(
                result['response_time'],
                result['status_code'],
                result['error']
            )
    
    metrics.end_time = time.time()
    stop_event.set()
    monitor_thread.join()
    
    return metrics.calculate_stats()

def print_performance_report(test_name, stats):
    """Print detailed performance report"""
    print(f"\n📊 {test_name} PERFORMANCE REPORT")
    print("=" * 80)
    
    print(f"🔢 REQUEST METRICS:")
    print(f"   Total Requests: {stats['total_requests']}")
    print(f"   Successful: {stats['successful_requests']} ({stats['success_rate']:.2f}%)")
    print(f"   Failed: {stats['failed_requests']} ({stats['error_rate']:.2f}%)")
    print(f"   Duration: {stats['total_duration']:.2f} seconds")
    print(f"   Requests/Second: {stats['requests_per_second']:.2f} RPS")
    
    print(f"\n⏱️  RESPONSE TIME METRICS:")
    print(f"   Average: {stats['avg_response_time']:.3f}s")
    print(f"   Median: {stats['median_response_time']:.3f}s")
    print(f"   Minimum: {stats['min_response_time']:.3f}s")
    print(f"   Maximum: {stats['max_response_time']:.3f}s")
    print(f"   95th Percentile: {stats['p95_response_time']:.3f}s")
    print(f"   99th Percentile: {stats['p99_response_time']:.3f}s")
    
    print(f"\n📈 HTTP STATUS CODES:")
    for status_code, count in stats['status_codes'].items():
        percentage = (count / stats['total_requests']) * 100
        print(f"   {status_code}: {count} ({percentage:.1f}%)")
    
    print(f"\n💻 SYSTEM RESOURCE USAGE:")
    print(f"   Average Memory: {stats['avg_memory_usage']:.1f}%")
    print(f"   Peak Memory: {stats['max_memory_usage']:.1f}%")
    print(f"   Average CPU: {stats['avg_cpu_usage']:.1f}%")
    print(f"   Peak CPU: {stats['max_cpu_usage']:.1f}%")

def run_comprehensive_performance_tests():
    """Run comprehensive high-load performance tests"""
    print("🚀 STARTING COMPREHENSIVE HIGH-LOAD PERFORMANCE TESTING")
    print("=" * 80)
    
    # Test scenarios with different request counts
    test_scenarios = [
        ("Health Check Lightweight", test_health_check_load, 1000),
        ("OTP Generation (DB Writes)", test_otp_generation_load, 500),
        ("Tournament Listing (DB Reads)", test_tournament_listing_load, 500),
        ("Tournament Creation (Complex DB Ops)", test_tournament_creation_load, 200),
        ("Mixed Workload Simulation", test_mixed_workload, 1000),
    ]
    
    all_results = {}
    
    for test_name, test_function, request_count in test_scenarios:
        try:
            print(f"\n🎯 Starting {test_name}...")
            stats = test_function(request_count)
            all_results[test_name] = stats
            print_performance_report(test_name, stats)
            
            # Brief pause between tests
            time.sleep(2)
            
        except Exception as e:
            print(f"❌ Error in {test_name}: {e}")
            all_results[test_name] = {"error": str(e)}
    
    # Overall summary
    print("\n" + "=" * 80)
    print("🏁 OVERALL PERFORMANCE SUMMARY")
    print("=" * 80)
    
    total_requests = 0
    total_successful = 0
    total_failed = 0
    avg_rps_list = []
    avg_response_times = []
    
    for test_name, stats in all_results.items():
        if "error" not in stats:
            total_requests += stats['total_requests']
            total_successful += stats['successful_requests']
            total_failed += stats['failed_requests']
            avg_rps_list.append(stats['requests_per_second'])
            avg_response_times.append(stats['avg_response_time'])
            
            print(f"\n{test_name}:")
            print(f"  ✅ Success Rate: {stats['success_rate']:.1f}%")
            print(f"  ⚡ RPS: {stats['requests_per_second']:.1f}")
            print(f"  ⏱️  Avg Response: {stats['avg_response_time']:.3f}s")
            print(f"  📊 95th Percentile: {stats['p95_response_time']:.3f}s")
        else:
            print(f"\n{test_name}: ❌ FAILED - {stats['error']}")
    
    if avg_rps_list:
        print(f"\n🎯 AGGREGATE METRICS:")
        print(f"   Total Requests Tested: {total_requests}")
        print(f"   Overall Success Rate: {(total_successful/total_requests)*100:.1f}%")
        print(f"   Average RPS Across Tests: {statistics.mean(avg_rps_list):.1f}")
        print(f"   Average Response Time: {statistics.mean(avg_response_times):.3f}s")
    
    # Performance assessment
    print(f"\n🔍 PERFORMANCE ASSESSMENT:")
    
    if avg_rps_list:
        max_rps = max(avg_rps_list)
        min_response = min(avg_response_times)
        overall_success_rate = (total_successful/total_requests)*100
        
        if max_rps > 500 and min_response < 0.1 and overall_success_rate > 95:
            print("   🟢 EXCELLENT: High throughput, low latency, high reliability")
        elif max_rps > 200 and min_response < 0.5 and overall_success_rate > 90:
            print("   🟡 GOOD: Acceptable performance under load")
        elif max_rps > 50 and overall_success_rate > 80:
            print("   🟠 FAIR: Performance needs optimization")
        else:
            print("   🔴 POOR: Significant performance issues detected")
    
    print(f"\n📋 RECOMMENDATIONS:")
    if avg_rps_list and max(avg_rps_list) < 100:
        print("   • Consider database connection pooling optimization")
        print("   • Review database query performance and indexing")
        print("   • Consider implementing caching for read operations")
    
    if avg_response_times and max(avg_response_times) > 1.0:
        print("   • Optimize slow database queries")
        print("   • Consider async processing for heavy operations")
        print("   • Review goroutine pool configuration")
    
    if total_failed > 0:
        print("   • Investigate error patterns and implement retry mechanisms")
        print("   • Review timeout configurations")
        print("   • Consider circuit breaker patterns for resilience")
    
    return all_results

if __name__ == "__main__":
    try:
        # Check if backend is running
        response = requests.get(f"{BACKEND_URL}/health", timeout=5)
        if response.status_code != 200:
            print(f"❌ Backend not responding properly. Status: {response.status_code}")
            sys.exit(1)
        
        print(f"✅ Backend is running and responsive")
        
        # Run comprehensive tests
        results = run_comprehensive_performance_tests()
        
        print(f"\n🎉 HIGH-LOAD PERFORMANCE TESTING COMPLETED")
        print("=" * 80)
        
    except requests.exceptions.RequestException as e:
        print(f"❌ Cannot connect to backend at {BACKEND_URL}: {e}")
        sys.exit(1)
    except KeyboardInterrupt:
        print(f"\n⚠️  Testing interrupted by user")
        sys.exit(1)
    except Exception as e:
        print(f"❌ Unexpected error during testing: {e}")
        sys.exit(1)