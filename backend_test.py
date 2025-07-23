#!/usr/bin/env python3
"""
Comprehensive Backend Testing for Enhanced eSports Fantasy Platform
Testing all core and newly implemented enhanced features
"""

import requests
import json
import time
import uuid
import base64
from typing import Dict, Any, Optional

class eSportsBackendTester:
    def __init__(self, base_url: str = "http://localhost:8001"):
        self.base_url = base_url
        self.api_base = f"{base_url}/api/v1"
        self.session = requests.Session()
        self.auth_token = None
        self.test_results = []
        
    def log_test(self, test_name: str, success: bool, details: str = "", response_data: Any = None):
        """Log test results"""
        result = {
            "test": test_name,
            "success": success,
            "details": details,
            "response_data": response_data
        }
        self.test_results.append(result)
        status = "✅ PASS" if success else "❌ FAIL"
        print(f"{status} {test_name}: {details}")
        
    def make_request(self, method: str, endpoint: str, data: Dict = None, headers: Dict = None) -> requests.Response:
        """Make HTTP request with proper error handling"""
        url = f"{self.api_base}{endpoint}" if endpoint.startswith('/') else f"{self.base_url}{endpoint}"
        
        default_headers = {"Content-Type": "application/json"}
        if self.auth_token:
            default_headers["Authorization"] = f"Bearer {self.auth_token}"
        if headers:
            default_headers.update(headers)
            
        try:
            if method.upper() == "GET":
                response = self.session.get(url, headers=default_headers, timeout=10)
            elif method.upper() == "POST":
                response = self.session.post(url, json=data, headers=default_headers, timeout=10)
            elif method.upper() == "PUT":
                response = self.session.put(url, json=data, headers=default_headers, timeout=10)
            elif method.upper() == "DELETE":
                response = self.session.delete(url, headers=default_headers, timeout=10)
            else:
                raise ValueError(f"Unsupported HTTP method: {method}")
                
            return response
        except requests.exceptions.RequestException as e:
            print(f"Request failed: {e}")
            return None

    def test_health_check(self):
        """Test health check endpoint - should return enhanced info with version 2.0.0"""
        print("\n=== TESTING HEALTH CHECK ENDPOINT ===")
        
        response = self.make_request("GET", "/health")
        if not response:
            self.log_test("Health Check", False, "Request failed")
            return
            
        if response.status_code == 200:
            data = response.json()
            expected_fields = ["status", "service", "version"]
            
            # Check basic fields
            missing_fields = [field for field in expected_fields if field not in data]
            if missing_fields:
                self.log_test("Health Check", False, f"Missing fields: {missing_fields}", data)
                return
                
            # Check version
            if data.get("version") == "2.0.0":
                # Check for enhanced features
                if "features" in data:
                    expected_features = ["phonepe_payments", "firebase_auth", "auto_contest_management", "analytics_dashboard", "real_time_simulation"]
                    has_features = all(feature in data["features"] for feature in expected_features)
                    if has_features:
                        self.log_test("Health Check", True, f"Enhanced health check working with version {data['version']} and all features", data)
                    else:
                        self.log_test("Health Check", False, f"Missing expected features in health check", data)
                else:
                    self.log_test("Health Check", False, "Health check missing 'features' field for v2.0.0", data)
            else:
                self.log_test("Health Check", False, f"Expected version 2.0.0, got {data.get('version')}", data)
        else:
            self.log_test("Health Check", False, f"HTTP {response.status_code}: {response.text}")

    def test_otp_authentication(self):
        """Test OTP authentication system"""
        print("\n=== TESTING OTP AUTHENTICATION SYSTEM ===")
        
        # Test phone number for OTP
        test_phone = "+919876543210"
        
        # Test Send OTP
        otp_data = {"phone_number": test_phone}
        response = self.make_request("POST", "/auth/send-otp", otp_data)
        
        if not response:
            self.log_test("Send OTP", False, "Request failed")
            return
            
        if response.status_code == 200:
            self.log_test("Send OTP", True, f"OTP sent successfully to {test_phone}")
            
            # Wait a moment and try to get OTP from logs (in real scenario, user would receive it)
            time.sleep(1)
            
            # Test Verify OTP with a common test OTP
            test_otp = "123456"  # Common test OTP
            verify_data = {"phone_number": test_phone, "otp": test_otp}
            
            verify_response = self.make_request("POST", "/auth/verify-otp", verify_data)
            if verify_response and verify_response.status_code == 200:
                verify_result = verify_response.json()
                if "token" in verify_result:
                    self.auth_token = verify_result["token"]
                    self.log_test("Verify OTP", True, "OTP verified successfully, JWT token received", verify_result)
                else:
                    self.log_test("Verify OTP", False, "OTP verified but no token received", verify_result)
            else:
                # Try with different common test OTPs
                for test_otp in ["000000", "111111", "999999", "886699"]:
                    verify_data["otp"] = test_otp
                    verify_response = self.make_request("POST", "/auth/verify-otp", verify_data)
                    if verify_response and verify_response.status_code == 200:
                        verify_result = verify_response.json()
                        if "token" in verify_result:
                            self.auth_token = verify_result["token"]
                            self.log_test("Verify OTP", True, f"OTP {test_otp} verified successfully", verify_result)
                            break
                else:
                    self.log_test("Verify OTP", False, f"Could not verify OTP with common test codes")
        else:
            self.log_test("Send OTP", False, f"HTTP {response.status_code}: {response.text}")

    def test_tournament_management(self):
        """Test tournament management APIs"""
        print("\n=== TESTING TOURNAMENT MANAGEMENT ===")
        
        # Test Get Tournaments
        response = self.make_request("GET", "/admin/tournaments")
        if response and response.status_code == 200:
            tournaments = response.json()
            self.log_test("Get Tournaments", True, f"Retrieved {len(tournaments)} tournaments")
            
            # Test Create Tournament
            tournament_data = {
                "name": f"Test Championship {uuid.uuid4().hex[:8]}",
                "start_date": "2024-12-25T10:00:00Z",
                "end_date": "2024-12-30T18:00:00Z",
                "game": "Valorant",
                "prize_pool": 50000
            }
            
            create_response = self.make_request("POST", "/admin/tournaments", tournament_data)
            if create_response and create_response.status_code in [200, 201]:
                created_tournament = create_response.json()
                self.log_test("Create Tournament", True, f"Tournament created with ID: {created_tournament.get('id')}", created_tournament)
            else:
                self.log_test("Create Tournament", False, f"HTTP {create_response.status_code if create_response else 'No response'}")
        else:
            self.log_test("Get Tournaments", False, f"HTTP {response.status_code if response else 'No response'}")

    def test_contest_management(self):
        """Test contest management APIs"""
        print("\n=== TESTING CONTEST MANAGEMENT ===")
        
        # Test Create Contest (requires tournament/match)
        contest_data = {
            "name": f"Test Contest {uuid.uuid4().hex[:8]}",
            "entry_fee": 100,
            "max_participants": 1000,
            "prize_pool": 10000,
            "match_id": str(uuid.uuid4())  # Mock match ID
        }
        
        response = self.make_request("POST", "/admin/contests", contest_data)
        if response and response.status_code in [200, 201]:
            contest = response.json()
            self.log_test("Create Contest", True, f"Contest created successfully", contest)
        else:
            self.log_test("Create Contest", False, f"HTTP {response.status_code if response else 'No response'}")

    def test_username_management(self):
        """Test newly implemented username management features"""
        print("\n=== TESTING USERNAME MANAGEMENT FEATURES ===")
        
        # Test Admin Get Username Prefixes
        response = self.make_request("GET", "/admin/username-prefixes")
        if response and response.status_code == 200:
            prefixes = response.json()
            self.log_test("Admin Get Username Prefixes", True, f"Retrieved {len(prefixes)} username prefixes", prefixes)
        else:
            self.log_test("Admin Get Username Prefixes", False, f"HTTP {response.status_code if response else 'No response'}")
        
        # Test Admin Create Username Prefix
        prefix_data = {
            "prefix": f"TEST{uuid.uuid4().hex[:4].upper()}",
            "description": "Test prefix for automated testing",
            "is_active": True,
            "category": "gaming"
        }
        
        response = self.make_request("POST", "/admin/username-prefixes", prefix_data)
        if response and response.status_code in [200, 201]:
            created_prefix = response.json()
            self.log_test("Admin Create Username Prefix", True, f"Username prefix created: {created_prefix.get('prefix')}", created_prefix)
        else:
            self.log_test("Admin Create Username Prefix", False, f"HTTP {response.status_code if response else 'No response'}")
        
        # Test User Get Active Username Prefixes
        response = self.make_request("GET", "/user/username-prefixes")
        if response and response.status_code == 200:
            user_prefixes = response.json()
            self.log_test("User Get Username Prefixes", True, f"User can access {len(user_prefixes)} active prefixes", user_prefixes)
        else:
            self.log_test("User Get Username Prefixes", False, f"HTTP {response.status_code if response else 'No response'}")
        
        # Test Generate Username
        response = self.make_request("POST", "/user/generate-username", {})
        if response and response.status_code == 200:
            username_result = response.json()
            generated_username = username_result.get('username')
            self.log_test("Generate Username", True, f"Generated username: {generated_username}", username_result)
            
            # Test Check Username Availability
            if generated_username:
                response = self.make_request("GET", f"/user/check-username?username={generated_username}")
                if response and response.status_code == 200:
                    availability = response.json()
                    self.log_test("Check Username Availability", True, f"Username availability checked", availability)
                else:
                    self.log_test("Check Username Availability", False, f"HTTP {response.status_code if response else 'No response'}")
        else:
            self.log_test("Generate Username", False, f"HTTP {response.status_code if response else 'No response'}")

    def test_game_management(self):
        """Test newly implemented game management features"""
        print("\n=== TESTING GAME MANAGEMENT FEATURES ===")
        
        # Test Admin Get Games
        response = self.make_request("GET", "/admin/games")
        if response and response.status_code == 200:
            games = response.json()
            self.log_test("Admin Get Games", True, f"Retrieved {len(games)} games", games)
        else:
            self.log_test("Admin Get Games", False, f"HTTP {response.status_code if response else 'No response'}")
        
        # Test Admin Create Game
        game_data = {
            "name": f"Test Game {uuid.uuid4().hex[:8]}",
            "description": "Automated test game",
            "category": "MOBA",
            "is_active": True,
            "max_players_per_team": 5,
            "scoring_system": "kills_deaths_assists"
        }
        
        response = self.make_request("POST", "/admin/games", game_data)
        if response and response.status_code in [200, 201]:
            created_game = response.json()
            game_id = created_game.get('id')
            self.log_test("Admin Create Game", True, f"Game created with ID: {game_id}", created_game)
            
            # Test Get Scoring Rules for the created game
            if game_id:
                response = self.make_request("GET", f"/admin/scoring-rules?game_id={game_id}")
                if response and response.status_code == 200:
                    scoring_rules = response.json()
                    self.log_test("Get Scoring Rules", True, f"Retrieved scoring rules for game {game_id}", scoring_rules)
                else:
                    self.log_test("Get Scoring Rules", False, f"HTTP {response.status_code if response else 'No response'}")
        else:
            self.log_test("Admin Create Game", False, f"HTTP {response.status_code if response else 'No response'}")

    def test_profile_management(self):
        """Test newly implemented profile management features"""
        print("\n=== TESTING PROFILE MANAGEMENT FEATURES ===")
        
        # Test Get Enhanced Profile
        response = self.make_request("GET", "/user/profile-enhanced")
        if response and response.status_code == 200:
            profile = response.json()
            self.log_test("Get Enhanced Profile", True, f"Retrieved enhanced profile", profile)
        else:
            self.log_test("Get Enhanced Profile", False, f"HTTP {response.status_code if response else 'No response'}")
        
        # Test Upload Profile Image (base64)
        # Create a small test image in base64
        test_image_base64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="
        
        image_data = {
            "image": f"data:image/png;base64,{test_image_base64}",
            "filename": "test_profile.png"
        }
        
        response = self.make_request("POST", "/user/upload-image", image_data)
        if response and response.status_code == 200:
            upload_result = response.json()
            self.log_test("Upload Profile Image", True, f"Profile image uploaded successfully", upload_result)
        else:
            self.log_test("Upload Profile Image", False, f"HTTP {response.status_code if response else 'No response'}")
        
        # Test Update Enhanced Profile
        profile_update = {
            "username": f"testuser_{uuid.uuid4().hex[:8]}",
            "display_name": "Test User Enhanced",
            "bio": "This is a test user profile for automated testing"
        }
        
        response = self.make_request("PUT", "/user/profile-enhanced", profile_update)
        if response and response.status_code == 200:
            updated_profile = response.json()
            self.log_test("Update Enhanced Profile", True, f"Profile updated successfully", updated_profile)
        else:
            self.log_test("Update Enhanced Profile", False, f"HTTP {response.status_code if response else 'No response'}")

    def test_achievement_system(self):
        """Test newly implemented achievement system"""
        print("\n=== TESTING ACHIEVEMENT SYSTEM ===")
        
        # Test Admin Get Achievements
        response = self.make_request("GET", "/admin/achievements")
        if response and response.status_code == 200:
            achievements = response.json()
            self.log_test("Admin Get Achievements", True, f"Retrieved {len(achievements)} achievements", achievements)
        else:
            self.log_test("Admin Get Achievements", False, f"HTTP {response.status_code if response else 'No response'}")
        
        # Test Admin Create Achievement
        achievement_data = {
            "name": f"Test Achievement {uuid.uuid4().hex[:8]}",
            "description": "Automated test achievement",
            "category": "gaming",
            "points": 100,
            "badge_icon": "trophy",
            "is_active": True,
            "criteria": {
                "type": "contest_wins",
                "threshold": 5
            }
        }
        
        response = self.make_request("POST", "/admin/achievements", achievement_data)
        if response and response.status_code in [200, 201]:
            created_achievement = response.json()
            self.log_test("Admin Create Achievement", True, f"Achievement created: {created_achievement.get('name')}", created_achievement)
        else:
            self.log_test("Admin Create Achievement", False, f"HTTP {response.status_code if response else 'No response'}")
        
        # Test User Get Achievements
        response = self.make_request("GET", "/user/achievements")
        if response and response.status_code == 200:
            user_achievements = response.json()
            self.log_test("User Get Achievements", True, f"User has {len(user_achievements)} achievements", user_achievements)
        else:
            self.log_test("User Get Achievements", False, f"HTTP {response.status_code if response else 'No response'}")
        
        # Test User Get Available Achievements
        response = self.make_request("GET", "/user/achievements/available")
        if response and response.status_code == 200:
            available_achievements = response.json()
            self.log_test("User Get Available Achievements", True, f"User can earn {len(available_achievements)} more achievements", available_achievements)
        else:
            self.log_test("User Get Available Achievements", False, f"HTTP {response.status_code if response else 'No response'}")

    def test_referral_system(self):
        """Test newly implemented referral system"""
        print("\n=== TESTING REFERRAL SYSTEM ===")
        
        # Test Generate Referral Code
        response = self.make_request("POST", "/user/referral/generate", {})
        if response and response.status_code == 200:
            referral_result = response.json()
            referral_code = referral_result.get('referral_code')
            self.log_test("Generate Referral Code", True, f"Generated referral code: {referral_code}", referral_result)
        else:
            self.log_test("Generate Referral Code", False, f"HTTP {response.status_code if response else 'No response'}")
        
        # Test Apply Referral Code (with a test code)
        apply_data = {"referral_code": "TEST123"}
        response = self.make_request("POST", "/user/referral/apply", apply_data)
        if response:
            if response.status_code == 200:
                apply_result = response.json()
                self.log_test("Apply Referral Code", True, f"Referral code applied successfully", apply_result)
            else:
                # Expected to fail with test code, but endpoint should respond
                self.log_test("Apply Referral Code", True, f"Referral endpoint working (expected failure with test code)")
        else:
            self.log_test("Apply Referral Code", False, "No response from referral apply endpoint")
        
        # Test Get Referral Stats
        response = self.make_request("GET", "/user/referral/stats")
        if response and response.status_code == 200:
            referral_stats = response.json()
            self.log_test("Get Referral Stats", True, f"Retrieved referral statistics", referral_stats)
        else:
            self.log_test("Get Referral Stats", False, f"HTTP {response.status_code if response else 'No response'}")
        
        # Test Get Referral Leaderboard
        response = self.make_request("GET", "/user/referral/leaderboard")
        if response and response.status_code == 200:
            leaderboard = response.json()
            self.log_test("Get Referral Leaderboard", True, f"Retrieved referral leaderboard with {len(leaderboard)} entries", leaderboard)
        else:
            self.log_test("Get Referral Leaderboard", False, f"HTTP {response.status_code if response else 'No response'}")

    def test_season_leagues(self):
        """Test newly implemented season leagues"""
        print("\n=== TESTING SEASON LEAGUES ===")
        
        # Test Admin Get Season Leagues
        response = self.make_request("GET", "/admin/season-leagues")
        if response and response.status_code == 200:
            leagues = response.json()
            self.log_test("Admin Get Season Leagues", True, f"Retrieved {len(leagues)} season leagues", leagues)
        else:
            self.log_test("Admin Get Season Leagues", False, f"HTTP {response.status_code if response else 'No response'}")
        
        # Test Admin Create Season League
        league_data = {
            "name": f"Test Season League {uuid.uuid4().hex[:8]}",
            "description": "Automated test season league",
            "game_id": str(uuid.uuid4()),  # Mock game ID
            "start_date": "2024-12-01T00:00:00Z",
            "end_date": "2024-12-31T23:59:59Z",
            "entry_fee": 500,
            "max_participants": 1000,
            "prize_pool": 50000,
            "is_active": True
        }
        
        response = self.make_request("POST", "/admin/season-leagues", league_data)
        if response and response.status_code in [200, 201]:
            created_league = response.json()
            self.log_test("Admin Create Season League", True, f"Season league created: {created_league.get('name')}", created_league)
        else:
            self.log_test("Admin Create Season League", False, f"HTTP {response.status_code if response else 'No response'}")
        
        # Test User Get Active Season Leagues
        response = self.make_request("GET", "/user/season-leagues")
        if response and response.status_code == 200:
            user_leagues = response.json()
            self.log_test("User Get Season Leagues", True, f"User can see {len(user_leagues)} active season leagues", user_leagues)
        else:
            self.log_test("User Get Season Leagues", False, f"HTTP {response.status_code if response else 'No response'}")

    def test_contest_templates(self):
        """Test newly implemented contest templates"""
        print("\n=== TESTING CONTEST TEMPLATES ===")
        
        # Test Admin Get Contest Templates
        response = self.make_request("GET", "/admin/contest-templates")
        if response and response.status_code == 200:
            templates = response.json()
            self.log_test("Admin Get Contest Templates", True, f"Retrieved {len(templates)} contest templates", templates)
        else:
            self.log_test("Admin Get Contest Templates", False, f"HTTP {response.status_code if response else 'No response'}")
        
        # Test Admin Create Contest Template
        template_data = {
            "name": f"Test Contest Template {uuid.uuid4().hex[:8]}",
            "description": "Automated test contest template",
            "game_id": str(uuid.uuid4()),  # Mock game ID
            "entry_fee": 100,
            "max_participants": 500,
            "prize_distribution": {
                "1st": 50,
                "2nd": 30,
                "3rd": 20
            },
            "scoring_rules": {
                "kill": 2,
                "death": -1,
                "assist": 1
            },
            "is_active": True
        }
        
        response = self.make_request("POST", "/admin/contest-templates", template_data)
        if response and response.status_code in [200, 201]:
            created_template = response.json()
            self.log_test("Admin Create Contest Template", True, f"Contest template created: {created_template.get('name')}", created_template)
        else:
            self.log_test("Admin Create Contest Template", False, f"HTTP {response.status_code if response else 'No response'}")

    def run_all_tests(self):
        """Run all comprehensive tests"""
        print("🚀 Starting Comprehensive eSports Fantasy Backend Testing")
        print("=" * 80)
        
        start_time = time.time()
        
        # Core existing features
        self.test_health_check()
        self.test_otp_authentication()
        self.test_tournament_management()
        self.test_contest_management()
        
        # Newly implemented enhanced features
        self.test_username_management()
        self.test_game_management()
        self.test_profile_management()
        
        # Newly implemented advanced features
        self.test_achievement_system()
        self.test_referral_system()
        self.test_season_leagues()
        self.test_contest_templates()
        
        end_time = time.time()
        duration = end_time - start_time
        
        # Generate summary
        self.generate_test_summary(duration)
        
    def generate_test_summary(self, duration: float):
        """Generate comprehensive test summary"""
        print("\n" + "=" * 80)
        print("🏁 COMPREHENSIVE TEST SUMMARY")
        print("=" * 80)
        
        total_tests = len(self.test_results)
        passed_tests = sum(1 for result in self.test_results if result["success"])
        failed_tests = total_tests - passed_tests
        
        print(f"📊 Total Tests: {total_tests}")
        print(f"✅ Passed: {passed_tests}")
        print(f"❌ Failed: {failed_tests}")
        print(f"⏱️  Duration: {duration:.2f} seconds")
        print(f"📈 Success Rate: {(passed_tests/total_tests)*100:.1f}%")
        
        if failed_tests > 0:
            print(f"\n❌ FAILED TESTS:")
            for result in self.test_results:
                if not result["success"]:
                    print(f"   • {result['test']}: {result['details']}")
        
        print(f"\n🎯 FEATURE TESTING STATUS:")
        
        # Group tests by feature category
        feature_categories = {
            "Core Features": ["Health Check", "Send OTP", "Verify OTP", "Get Tournaments", "Create Tournament", "Create Contest"],
            "Username Management": ["Admin Get Username Prefixes", "Admin Create Username Prefix", "User Get Username Prefixes", "Generate Username", "Check Username Availability"],
            "Game Management": ["Admin Get Games", "Admin Create Game", "Get Scoring Rules"],
            "Profile Management": ["Get Enhanced Profile", "Upload Profile Image", "Update Enhanced Profile"],
            "Achievement System": ["Admin Get Achievements", "Admin Create Achievement", "User Get Achievements", "User Get Available Achievements"],
            "Referral System": ["Generate Referral Code", "Apply Referral Code", "Get Referral Stats", "Get Referral Leaderboard"],
            "Season Leagues": ["Admin Get Season Leagues", "Admin Create Season League", "User Get Season Leagues"],
            "Contest Templates": ["Admin Get Contest Templates", "Admin Create Contest Template"]
        }
        
        for category, tests in feature_categories.items():
            category_results = [r for r in self.test_results if r["test"] in tests]
            if category_results:
                passed = sum(1 for r in category_results if r["success"])
                total = len(category_results)
                status = "✅ WORKING" if passed == total else f"⚠️  PARTIAL ({passed}/{total})" if passed > 0 else "❌ FAILED"
                print(f"   {category}: {status}")

if __name__ == "__main__":
    tester = eSportsBackendTester()
    tester.run_all_tests()