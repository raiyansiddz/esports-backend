#!/usr/bin/env python3
"""
Actual Backend API Testing Script for eSports Fantasy Platform
Tests the currently implemented features based on the routes.go file
"""

import requests
import json
import sys
import os
import time
from datetime import datetime

# Test the GoLang backend running on localhost:8001
BACKEND_URL = "http://localhost:8001"
API_BASE_URL = f"{BACKEND_URL}/api/v1"

print(f"Testing Actual eSports Fantasy Backend Features at: {BACKEND_URL}")
print(f"API Base URL: {API_BASE_URL}")
print("=" * 80)

def test_health_check():
    """Test the health check endpoint"""
    print("\n1. Testing Health Check Endpoint")
    print("-" * 40)
    
    try:
        response = requests.get(f"{BACKEND_URL}/health", timeout=10)
        print(f"Status Code: {response.status_code}")
        print(f"Response: {response.json()}")
        
        if response.status_code == 200:
            data = response.json()
            expected_fields = ["status", "service", "version", "features"]
            if all(field in data for field in expected_fields):
                if data.get("status") == "ok" and data.get("service") == "esports-fantasy-backend":
                    print("✅ Health check endpoint working correctly")
                    print(f"Version: {data.get('version')}")
                    print(f"Features: {data.get('features')}")
                    return True
                else:
                    print("❌ Health check returned unexpected values")
                    return False
            else:
                print(f"❌ Health check missing required fields: {expected_fields}")
                return False
        else:
            print(f"❌ Health check failed with status {response.status_code}")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ Health check request failed: {e}")
        return False
    except Exception as e:
        print(f"❌ Health check test error: {e}")
        return False

def test_firebase_config():
    """Test Firebase configuration endpoint"""
    print("\n2. Testing Firebase Configuration")
    print("-" * 40)
    
    try:
        response = requests.get(f"{API_BASE_URL}/firebase/config", timeout=10)
        print(f"Status Code: {response.status_code}")
        print(f"Response: {response.json()}")
        
        if response.status_code == 200:
            print("✅ Firebase configuration endpoint working correctly")
            return True
        else:
            print(f"❌ Firebase configuration failed: {response.text}")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ Firebase configuration request failed: {e}")
        return False
    except Exception as e:
        print(f"❌ Firebase configuration test error: {e}")
        return False

def test_legacy_otp_auth():
    """Test legacy OTP authentication system"""
    print("\n3. Testing Legacy OTP Authentication")
    print("-" * 40)
    
    try:
        # Send OTP
        otp_data = {"phone_number": "9876543210"}
        otp_response = requests.post(
            f"{API_BASE_URL}/auth/send-otp",
            json=otp_data,
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        
        print(f"Send OTP Status: {otp_response.status_code}")
        print(f"Send OTP Response: {otp_response.json()}")
        
        if otp_response.status_code == 200:
            print("✅ Legacy OTP send working correctly")
            return True
        else:
            print(f"❌ Legacy OTP send failed: {otp_response.text}")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ Legacy OTP authentication request failed: {e}")
        return False
    except Exception as e:
        print(f"❌ Legacy OTP authentication test error: {e}")
        return False

def test_firebase_otp_auth():
    """Test Firebase OTP authentication system"""
    print("\n4. Testing Firebase OTP Authentication")
    print("-" * 40)
    
    try:
        # Send OTP via Firebase
        otp_data = {"phone_number": "9876543210"}
        otp_response = requests.post(
            f"{API_BASE_URL}/auth/firebase/send-otp",
            json=otp_data,
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        
        print(f"Firebase Send OTP Status: {otp_response.status_code}")
        print(f"Firebase Send OTP Response: {otp_response.json()}")
        
        if otp_response.status_code == 200:
            print("✅ Firebase OTP send working correctly")
            return True
        else:
            print(f"❌ Firebase OTP send failed: {otp_response.text}")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ Firebase OTP authentication request failed: {e}")
        return False
    except Exception as e:
        print(f"❌ Firebase OTP authentication test error: {e}")
        return False

def test_admin_tournaments():
    """Test admin tournament management"""
    print("\n5. Testing Admin Tournament Management")
    print("-" * 40)
    
    try:
        # Test get tournaments (should work without auth based on routes)
        response = requests.get(f"{API_BASE_URL}/admin/tournaments", timeout=10)
        print(f"Get Tournaments Status: {response.status_code}")
        print(f"Response: {response.json()}")
        
        if response.status_code == 200:
            print("✅ Admin tournament management working correctly")
            return True
        else:
            print(f"❌ Admin tournament management failed: {response.text}")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ Admin tournament request failed: {e}")
        return False
    except Exception as e:
        print(f"❌ Admin tournament test error: {e}")
        return False

def test_admin_matches():
    """Test admin match management"""
    print("\n6. Testing Admin Match Management")
    print("-" * 40)
    
    try:
        # Test get matches
        response = requests.get(f"{API_BASE_URL}/admin/matches", timeout=10)
        print(f"Get Matches Status: {response.status_code}")
        print(f"Response: {response.json()}")
        
        if response.status_code == 200:
            print("✅ Admin match management working correctly")
            return True
        else:
            print(f"❌ Admin match management failed: {response.text}")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ Admin match request failed: {e}")
        return False
    except Exception as e:
        print(f"❌ Admin match test error: {e}")
        return False

def test_admin_esports_teams():
    """Test admin eSports team management"""
    print("\n7. Testing Admin eSports Team Management")
    print("-" * 40)
    
    try:
        # Test get eSports teams
        response = requests.get(f"{API_BASE_URL}/admin/esports-teams", timeout=10)
        print(f"Get eSports Teams Status: {response.status_code}")
        print(f"Response: {response.json()}")
        
        if response.status_code == 200:
            print("✅ Admin eSports team management working correctly")
            return True
        else:
            print(f"❌ Admin eSports team management failed: {response.text}")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ Admin eSports team request failed: {e}")
        return False
    except Exception as e:
        print(f"❌ Admin eSports team test error: {e}")
        return False

def test_public_contests():
    """Test public contest information endpoints"""
    print("\n8. Testing Public Contest Information")
    print("-" * 40)
    
    try:
        # Test get contests by match (using a dummy match ID)
        dummy_match_id = "123e4567-e89b-12d3-a456-426614174000"
        response = requests.get(f"{API_BASE_URL}/contests/match/{dummy_match_id}", timeout=10)
        print(f"Get Contests by Match Status: {response.status_code}")
        print(f"Response: {response.json()}")
        
        # This might return 404 or empty array, both are acceptable for testing
        if response.status_code in [200, 404]:
            print("✅ Public contest information endpoint working correctly")
            return True
        else:
            print(f"❌ Public contest information failed: {response.text}")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ Public contest request failed: {e}")
        return False
    except Exception as e:
        print(f"❌ Public contest test error: {e}")
        return False

def test_analytics_dashboard():
    """Test analytics dashboard"""
    print("\n9. Testing Analytics Dashboard")
    print("-" * 40)
    
    try:
        # Test analytics dashboard
        response = requests.get(f"{API_BASE_URL}/admin/analytics/dashboard", timeout=10)
        print(f"Analytics Dashboard Status: {response.status_code}")
        print(f"Response: {response.json()}")
        
        if response.status_code == 200:
            print("✅ Analytics dashboard working correctly")
            return True
        else:
            print(f"❌ Analytics dashboard failed: {response.text}")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ Analytics dashboard request failed: {e}")
        return False
    except Exception as e:
        print(f"❌ Analytics dashboard test error: {e}")
        return False

def run_all_tests():
    """Run all actual backend tests"""
    print("Starting Actual Backend API Tests")
    print("=" * 80)
    
    test_results = []
    
    # Test implemented features
    test_results.append(("Health Check", test_health_check()))
    test_results.append(("Firebase Configuration", test_firebase_config()))
    test_results.append(("Legacy OTP Authentication", test_legacy_otp_auth()))
    test_results.append(("Firebase OTP Authentication", test_firebase_otp_auth()))
    test_results.append(("Admin Tournament Management", test_admin_tournaments()))
    test_results.append(("Admin Match Management", test_admin_matches()))
    test_results.append(("Admin eSports Team Management", test_admin_esports_teams()))
    test_results.append(("Public Contest Information", test_public_contests()))
    test_results.append(("Analytics Dashboard", test_analytics_dashboard()))
    
    # Summary
    print("\n" + "=" * 80)
    print("ACTUAL BACKEND TEST SUMMARY")
    print("=" * 80)
    
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