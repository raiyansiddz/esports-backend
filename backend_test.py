#!/usr/bin/env python3
"""
Backend API Testing Script for FastAPI MongoDB Application
Tests the actual implemented endpoints and functionality
"""

import requests
import json
import sys
import os
from datetime import datetime

# Get backend URL from frontend .env file
def get_backend_url():
    try:
        with open('/app/frontend/.env', 'r') as f:
            for line in f:
                if line.startswith('REACT_APP_BACKEND_URL='):
                    return line.split('=', 1)[1].strip()
    except Exception as e:
        print(f"Error reading frontend .env: {e}")
        return None

BACKEND_URL = get_backend_url()
if not BACKEND_URL:
    print("ERROR: Could not get REACT_APP_BACKEND_URL from frontend/.env")
    sys.exit(1)

API_BASE_URL = f"{BACKEND_URL}/api"

print(f"Testing Backend API at: {API_BASE_URL}")
print("=" * 60)

def test_root_endpoint():
    """Test the root API endpoint"""
    print("\n1. Testing Root Endpoint")
    print("-" * 30)
    
    try:
        response = requests.get(f"{API_BASE_URL}/", timeout=10)
        print(f"Status Code: {response.status_code}")
        print(f"Response: {response.json()}")
        
        if response.status_code == 200:
            data = response.json()
            if data.get("message") == "Hello World":
                print("✅ Root endpoint working correctly")
                return True
            else:
                print("❌ Root endpoint returned unexpected message")
                return False
        else:
            print(f"❌ Root endpoint failed with status {response.status_code}")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ Root endpoint request failed: {e}")
        return False
    except Exception as e:
        print(f"❌ Root endpoint test error: {e}")
        return False

def test_create_status_check():
    """Test creating a status check"""
    print("\n2. Testing Create Status Check")
    print("-" * 30)
    
    test_data = {
        "client_name": "test_client_backend_api"
    }
    
    try:
        response = requests.post(
            f"{API_BASE_URL}/status", 
            json=test_data,
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        
        print(f"Status Code: {response.status_code}")
        print(f"Response: {response.json()}")
        
        if response.status_code == 200:
            data = response.json()
            required_fields = ["id", "client_name", "timestamp"]
            
            if all(field in data for field in required_fields):
                if data["client_name"] == test_data["client_name"]:
                    print("✅ Create status check working correctly")
                    return True, data["id"]
                else:
                    print("❌ Create status check returned wrong client_name")
                    return False, None
            else:
                print(f"❌ Create status check missing required fields: {required_fields}")
                return False, None
        else:
            print(f"❌ Create status check failed with status {response.status_code}")
            return False, None
            
    except requests.exceptions.RequestException as e:
        print(f"❌ Create status check request failed: {e}")
        return False, None
    except Exception as e:
        print(f"❌ Create status check test error: {e}")
        return False, None

def test_get_status_checks():
    """Test getting all status checks"""
    print("\n3. Testing Get Status Checks")
    print("-" * 30)
    
    try:
        response = requests.get(f"{API_BASE_URL}/status", timeout=10)
        print(f"Status Code: {response.status_code}")
        
        if response.status_code == 200:
            data = response.json()
            print(f"Number of status checks returned: {len(data)}")
            
            if isinstance(data, list):
                if len(data) > 0:
                    # Check if the first item has required fields
                    first_item = data[0]
                    required_fields = ["id", "client_name", "timestamp"]
                    
                    if all(field in first_item for field in required_fields):
                        print("✅ Get status checks working correctly")
                        print(f"Sample record: {first_item}")
                        return True
                    else:
                        print(f"❌ Status check records missing required fields: {required_fields}")
                        return False
                else:
                    print("✅ Get status checks working (empty list)")
                    return True
            else:
                print("❌ Get status checks should return a list")
                return False
        else:
            print(f"❌ Get status checks failed with status {response.status_code}")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ Get status checks request failed: {e}")
        return False
    except Exception as e:
        print(f"❌ Get status checks test error: {e}")
        return False

def test_database_persistence():
    """Test that data persists in database"""
    print("\n4. Testing Database Persistence")
    print("-" * 30)
    
    # Create a unique test record
    unique_client_name = f"persistence_test_{datetime.now().strftime('%Y%m%d_%H%M%S')}"
    test_data = {"client_name": unique_client_name}
    
    try:
        # Create the record
        create_response = requests.post(
            f"{API_BASE_URL}/status", 
            json=test_data,
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        
        if create_response.status_code != 200:
            print(f"❌ Failed to create test record: {create_response.status_code}")
            return False
        
        created_record = create_response.json()
        created_id = created_record["id"]
        
        # Retrieve all records and check if our record exists
        get_response = requests.get(f"{API_BASE_URL}/status", timeout=10)
        
        if get_response.status_code != 200:
            print(f"❌ Failed to retrieve records: {get_response.status_code}")
            return False
        
        all_records = get_response.json()
        
        # Look for our record
        found_record = None
        for record in all_records:
            if record.get("id") == created_id and record.get("client_name") == unique_client_name:
                found_record = record
                break
        
        if found_record:
            print("✅ Database persistence working correctly")
            print(f"Created and retrieved record: {found_record}")
            return True
        else:
            print("❌ Created record not found in database")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ Database persistence test request failed: {e}")
        return False
    except Exception as e:
        print(f"❌ Database persistence test error: {e}")
        return False

def test_cors_headers():
    """Test CORS headers are present"""
    print("\n5. Testing CORS Headers")
    print("-" * 30)
    
    try:
        response = requests.get(f"{API_BASE_URL}/", timeout=10)
        
        cors_headers = {
            'access-control-allow-origin': response.headers.get('access-control-allow-origin'),
            'access-control-allow-methods': response.headers.get('access-control-allow-methods'),
            'access-control-allow-headers': response.headers.get('access-control-allow-headers'),
        }
        
        print("CORS Headers:")
        for header, value in cors_headers.items():
            print(f"  {header}: {value}")
        
        if cors_headers['access-control-allow-origin']:
            print("✅ CORS headers present")
            return True
        else:
            print("❌ CORS headers missing")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ CORS test request failed: {e}")
        return False
    except Exception as e:
        print(f"❌ CORS test error: {e}")
        return False

def test_error_handling():
    """Test error handling for invalid requests"""
    print("\n6. Testing Error Handling")
    print("-" * 30)
    
    try:
        # Test invalid JSON for POST request
        response = requests.post(
            f"{API_BASE_URL}/status", 
            data="invalid json",
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        
        print(f"Invalid JSON Status Code: {response.status_code}")
        
        if response.status_code in [400, 422]:  # FastAPI returns 422 for validation errors
            print("✅ Error handling working for invalid JSON")
            return True
        else:
            print(f"❌ Expected 400/422 for invalid JSON, got {response.status_code}")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ Error handling test request failed: {e}")
        return False
    except Exception as e:
        print(f"❌ Error handling test error: {e}")
        return False

def run_all_tests():
    """Run all backend tests"""
    print("Starting Backend API Tests")
    print("=" * 60)
    
    test_results = []
    
    # Run all tests
    test_results.append(("Root Endpoint", test_root_endpoint()))
    test_results.append(("Create Status Check", test_create_status_check()[0]))
    test_results.append(("Get Status Checks", test_get_status_checks()))
    test_results.append(("Database Persistence", test_database_persistence()))
    test_results.append(("CORS Headers", test_cors_headers()))
    test_results.append(("Error Handling", test_error_handling()))
    
    # Summary
    print("\n" + "=" * 60)
    print("TEST SUMMARY")
    print("=" * 60)
    
    passed = 0
    failed = 0
    
    for test_name, result in test_results:
        status = "✅ PASS" if result else "❌ FAIL"
        print(f"{test_name}: {status}")
        if result:
            passed += 1
        else:
            failed += 1
    
    print(f"\nTotal Tests: {len(test_results)}")
    print(f"Passed: {passed}")
    print(f"Failed: {failed}")
    
    if failed == 0:
        print("\n🎉 All tests passed! Backend API is working correctly.")
        return True
    else:
        print(f"\n⚠️  {failed} test(s) failed. Backend needs attention.")
        return False

if __name__ == "__main__":
    success = run_all_tests()
    sys.exit(0 if success else 1)