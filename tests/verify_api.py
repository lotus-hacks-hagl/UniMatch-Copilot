import requests
import json
import socket
import time

BASE_URL = 'http://localhost:8894/api/v1'
print("=== UniMatch API Integration Test ===")

def check_port(port):
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    result = sock.connect_ex(('localhost', port))
    sock.close()
    return result == 0

if not check_port(8894):
    print("Backend (Port 8894) is not running!")
    exit(1)

# 1. Test Auth (Register -> Login)
print("\n[1] Testing Authentication Flow...")
try:
    resp_reg = requests.post(f"{BASE_URL}/auth/register", json={"username": "admin_test", "password": "password123"}, timeout=5)
    print(f"Register Response: {resp_reg.status_code}")
    
    resp = requests.post(f"{BASE_URL}/auth/login", json={"username": "admin_test", "password": "password123"}, timeout=5)
    print(f"Login Response: {resp.status_code}")
    if resp.status_code == 200:
        token = resp.json().get('data', {}).get('token')
        print(f"Token acquired. Length: {len(token) if token else 0}")
    else:
        print(f"Login Failed via admin1: {resp.text}")
        token = None
except Exception as e:
    print(f"Auth Error: {e}")
    token = None

if token:
    headers = {"Authorization": f"Bearer {token}"}
    
    # 2. Test Get Cases
    print("\n[2] Testing GET /cases...")
    resp = requests.get(f"{BASE_URL}/cases?status=all_cases", headers=headers)
    print(f"Dashboard Cases: {resp.status_code}")
    
    # 3. Test Dashboard Stats
    print("\n[3] Testing GET /cases/stats...")
    resp = requests.get(f"{BASE_URL}/cases/stats", headers=headers)
    print(f"Dashboard Stats: {resp.status_code} | {resp.text[:100]}")
    
    # 4. Test Universities KB
    print("\n[4] Testing GET /universities...")
    resp = requests.get(f"{BASE_URL}/universities?page=1&limit=5", headers=headers)
    print(f"Uni KB list: {resp.status_code}")
else:
    print("Skipping authenticated endpoints due to login failure.")
