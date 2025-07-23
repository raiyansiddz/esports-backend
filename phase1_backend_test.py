#!/usr/bin/env python3
"""
Phase 1 Backend API Testing Script for Enhanced eSports Fantasy Platform
Tests the newly implemented Phase 1 features including:
1. Enhanced User Management (username generation, profile images)
2. Admin Game Management
3. Database Schema Updates
"""

import requests
import json
import sys
import os
import time
import base64
from datetime import datetime

# Test the Python FastAPI backend running on localhost:8001
BACKEND_URL = "http://localhost:8001"
API_BASE_URL = f"{BACKEND_URL}/api/v1"

print(f"Testing Enhanced eSports Fantasy Backend Phase 1 Features at: {BACKEND_URL}")
print(f"API Base URL: {API_BASE_URL}")
print("=" * 80)

# Global variables for authentication
auth_token = None
admin_token = None
test_user_phone = "9876543210"
admin_phone = "9999999999"

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
            expected_fields = ["service", "status", "version"]
            if all(field in data for field in expected_fields):
                if data.get("status") == "ok" and data.get("service") == "esports-fantasy-backend":
                    print("✅ Health check endpoint working correctly")
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

def authenticate_user(phone_number, is_admin=False):
    """Authenticate user and get JWT token"""
    print(f"\n2. Authenticating {'Admin' if is_admin else 'User'}: {phone_number}")
    print("-" * 40)
    
    try:
        # Send OTP
        otp_data = {"phone_number": phone_number}
        otp_response = requests.post(
            f"{API_BASE_URL}/auth/send-otp",
            json=otp_data,
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        
        print(f"Send OTP Status: {otp_response.status_code}")
        if otp_response.status_code != 200:
            print(f"❌ Failed to send OTP: {otp_response.text}")
            return None
        
        print("Note: Check backend logs for actual OTP in development mode")
        
        # Try common test OTPs that might be configured
        test_otps = ["123456", "000000", "111111", "999999"]
        
        for test_otp in test_otps:
            # Verify OTP
            verify_data = {"phone_number": phone_number, "otp": test_otp}
            verify_response = requests.post(
                f"{API_BASE_URL}/auth/verify-otp",
                json=verify_data,
                headers={"Content-Type": "application/json"},
                timeout=10
            )
            
            print(f"Trying OTP {test_otp}: Status {verify_response.status_code}")
            if verify_response.status_code == 200:
                data = verify_response.json()
                token = data.get("token")
                if token:
                    print(f"✅ Authentication successful for {'Admin' if is_admin else 'User'} with OTP {test_otp}")
                    return token
                else:
                    print("❌ No token received in response")
                    return None
        
        print(f"❌ All test OTPs failed. Need to check backend logs for actual OTP")
        return None
            
    except requests.exceptions.RequestException as e:
        print(f"❌ Authentication request failed: {e}")
        return None
    except Exception as e:
        print(f"❌ Authentication error: {e}")
        return None

def test_username_generation():
    """Test username generation system with admin-controlled prefixes"""
    print("\n3. Testing Username Generation System")
    print("-" * 40)
    
    if not auth_token:
        print("❌ No auth token available for testing")
        return False
    
    try:
        headers = {"Authorization": f"Bearer {auth_token}"}
        
        # First get available prefixes
        prefixes_response = requests.get(
            f"{API_BASE_URL}/user/username-prefixes",
            headers=headers,
            timeout=10
        )
        
        print(f"Get Prefixes Status: {prefixes_response.status_code}")
        if prefixes_response.status_code != 200:
            print(f"❌ Failed to get prefixes: {prefixes_response.text}")
            return False
        
        prefixes_data = prefixes_response.json()
        print(f"Available prefixes: {prefixes_data}")
        
        # Check if we have any active prefixes
        if "prefixes" not in prefixes_data or len(prefixes_data["prefixes"]) == 0:
            print("❌ No active prefixes available for username generation")
            return False
        
        # Use the first available prefix
        first_prefix = prefixes_data["prefixes"][0]
        prefix_id = first_prefix["id"]
        
        # Test generate username endpoint
        username_data = {"prefix_id": prefix_id}
        response = requests.post(
            f"{API_BASE_URL}/user/generate-username",
            json=username_data,
            headers=headers,
            timeout=10
        )
        
        print(f"Generate Username Status: {response.status_code}")
        print(f"Response: {response.json()}")
        
        if response.status_code == 200:
            data = response.json()
            if "username" in data:
                username = data["username"]
                print(f"Generated username: {username}")
                print("✅ Username generation working correctly")
                return True
            else:
                print("❌ No username in response")
                return False
        else:
            print(f"❌ Username generation failed: {response.text}")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ Username generation request failed: {e}")
        return False
    except Exception as e:
        print(f"❌ Username generation test error: {e}")
        return False

def test_profile_image_upload():
    """Test profile image upload with base64 storage"""
    print("\n4. Testing Profile Image Upload")
    print("-" * 40)
    
    if not auth_token:
        print("❌ No auth token available for testing")
        return False
    
    try:
        headers = {"Authorization": f"Bearer {auth_token}"}
        
        # Create a simple test image in base64 format (1x1 PNG)
        test_image_base64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChAI9jU8j8wAAAABJRU5ErkJggg=="
        
        upload_data = {
            "image": f"data:image/png;base64,{test_image_base64}"
        }
        
        response = requests.post(
            f"{API_BASE_URL}/user/upload-image",
            json=upload_data,
            headers=headers,
            timeout=10
        )
        
        print(f"Upload Image Status: {response.status_code}")
        print(f"Response: {response.json()}")
        
        if response.status_code == 200:
            data = response.json()
            if "message" in data or "profile_image" in data:
                print("✅ Profile image upload working correctly")
                return True
            else:
                print("❌ Unexpected response format")
                return False
        else:
            print(f"❌ Profile image upload failed: {response.text}")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ Profile image upload request failed: {e}")
        return False
    except Exception as e:
        print(f"❌ Profile image upload test error: {e}")
        return False

def test_user_profile_management():
    """Test enhanced profile management"""
    print("\n5. Testing Enhanced Profile Management")
    print("-" * 40)
    
    if not auth_token:
        print("❌ No auth token available for testing")
        return False
    
    try:
        headers = {"Authorization": f"Bearer {auth_token}"}
        
        # Get user profile
        response = requests.get(
            f"{API_BASE_URL}/user/profile",
            headers=headers,
            timeout=10
        )
        
        print(f"Get Profile Status: {response.status_code}")
        print(f"Response: {response.json()}")
        
        if response.status_code == 200:
            data = response.json()
            expected_fields = ["id", "phone_number"]
            if any(field in data for field in expected_fields):
                print("✅ Profile management working correctly")
                return True
            else:
                print(f"❌ Profile missing expected fields")
                return False
        else:
            print(f"❌ Profile retrieval failed: {response.text}")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ Profile management request failed: {e}")
        return False
    except Exception as e:
        print(f"❌ Profile management test error: {e}")
        return False

def test_admin_username_prefixes():
    """Test admin username prefix management"""
    print("\n6. Testing Admin Username Prefix Management")
    print("-" * 40)
    
    if not admin_token:
        print("❌ No admin token available for testing")
        return False
    
    try:
        headers = {"Authorization": f"Bearer {admin_token}"}
        
        # Test get username prefixes
        response = requests.get(
            f"{API_BASE_URL}/admin/username-prefixes",
            headers=headers,
            timeout=10
        )
        
        print(f"Get Username Prefixes Status: {response.status_code}")
        print(f"Response: {response.json()}")
        
        if response.status_code == 200:
            print("✅ Admin username prefix management working correctly")
            return True
        else:
            print(f"❌ Admin username prefix management failed: {response.text}")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ Admin username prefix request failed: {e}")
        return False
    except Exception as e:
        print(f"❌ Admin username prefix test error: {e}")
        return False

def test_admin_game_management():
    """Test admin game management system"""
    print("\n7. Testing Admin Game Management")
    print("-" * 40)
    
    if not admin_token:
        print("❌ No admin token available for testing")
        return False
    
    try:
        headers = {"Authorization": f"Bearer {admin_token}"}
        
        # Test get games
        response = requests.get(
            f"{API_BASE_URL}/admin/games",
            headers=headers,
            timeout=10
        )
        
        print(f"Get Games Status: {response.status_code}")
        print(f"Response: {response.json()}")
        
        if response.status_code == 200:
            # Test create game
            game_data = {
                "name": "Test Game",
                "description": "Test game for API testing",
                "max_team_size": 5,
                "player_roles": ["Assault", "Support", "Sniper"]
            }
            
            create_response = requests.post(
                f"{API_BASE_URL}/admin/games",
                json=game_data,
                headers=headers,
                timeout=10
            )
            
            print(f"Create Game Status: {create_response.status_code}")
            print(f"Create Response: {create_response.json()}")
            
            if create_response.status_code in [200, 201]:
                print("✅ Admin game management working correctly")
                return True
            else:
                print(f"❌ Game creation failed: {create_response.text}")
                return False
        else:
            print(f"❌ Admin game management failed: {response.text}")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ Admin game management request failed: {e}")
        return False
    except Exception as e:
        print(f"❌ Admin game management test error: {e}")
        return False

def test_admin_scoring_rules():
    """Test admin scoring rules management"""
    print("\n8. Testing Admin Scoring Rules Management")
    print("-" * 40)
    
    if not admin_token:
        print("❌ No admin token available for testing")
        return False
    
    try:
        headers = {"Authorization": f"Bearer {admin_token}"}
        
        # First get games to use for scoring rules
        games_response = requests.get(
            f"{API_BASE_URL}/admin/games",
            headers=headers,
            timeout=10
        )
        
        if games_response.status_code != 200:
            print(f"❌ Failed to get games for scoring rules test: {games_response.text}")
            return False
        
        games_data = games_response.json()
        if "games" not in games_data or len(games_data["games"]) == 0:
            print("❌ No games available for scoring rules test")
            return False
        
        # Use the first game
        first_game = games_data["games"][0]
        game_id = first_game["id"]
        
        # Test get scoring rules for this game
        response = requests.get(
            f"{API_BASE_URL}/admin/scoring-rules?game_id={game_id}",
            headers=headers,
            timeout=10
        )
        
        print(f"Get Scoring Rules Status: {response.status_code}")
        print(f"Response: {response.json()}")
        
        if response.status_code == 200:
            print("✅ Admin scoring rules management working correctly")
            return True
        else:
            print(f"❌ Admin scoring rules management failed: {response.text}")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ Admin scoring rules request failed: {e}")
        return False
    except Exception as e:
        print(f"❌ Admin scoring rules test error: {e}")
        return False

def run_all_tests():
    """Run all Phase 1 backend tests"""
    global auth_token, admin_token
    
    print("Starting Phase 1 Backend API Tests")
    print("=" * 80)
    
    test_results = []
    
    # Test 1: Health Check
    test_results.append(("Health Check", test_health_check()))
    
    # Test 2: User Authentication
    auth_token = authenticate_user(test_user_phone, is_admin=False)
    test_results.append(("User Authentication", auth_token is not None))
    
    # Test 3: Admin Authentication
    admin_token = authenticate_user(admin_phone, is_admin=True)
    test_results.append(("Admin Authentication", admin_token is not None))
    
    # Test 4: Username Generation
    test_results.append(("Username Generation", test_username_generation()))
    
    # Test 5: Profile Image Upload
    test_results.append(("Profile Image Upload", test_profile_image_upload()))
    
    # Test 6: User Profile Management
    test_results.append(("User Profile Management", test_user_profile_management()))
    
    # Test 7: Admin Username Prefixes
    test_results.append(("Admin Username Prefixes", test_admin_username_prefixes()))
    
    # Test 8: Admin Game Management
    test_results.append(("Admin Game Management", test_admin_game_management()))
    
    # Test 9: Admin Scoring Rules
    test_results.append(("Admin Scoring Rules", test_admin_scoring_rules()))
    
    # Summary
    print("\n" + "=" * 80)
    print("PHASE 1 TEST SUMMARY")
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
        print("\n🎉 All Phase 1 tests passed! Backend API is working correctly.")
        return True
    else:
        print(f"\n⚠️  {failed} test(s) failed. Backend needs attention.")
        return False

if __name__ == "__main__":
    success = run_all_tests()
    sys.exit(0 if success else 1)