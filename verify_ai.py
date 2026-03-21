import requests
import time

BASE_URL = 'http://localhost:8080/api/v1'

def run_ai_test():
    print("=== AI Engine Mock Test ===")
    
    # 1. Login
    resp = requests.post(f"{BASE_URL}/auth/login", json={"username": "admin_test", "password": "password123"})
    if resp.status_code != 200:
        print("Login Failed")
        return
    token = resp.json()['data']['token']
    headers = {"Authorization": f"Bearer {token}"}
    
    # 2. Create Case (Matching Flat CreateCaseRequest format)
    payload = {
        "full_name": "AI Verification Subject",
        "gpa_normalized": 3.75,
        "gpa_raw": 8.5,
        "gpa_scale": 10.0,
        "ielts_overall": 7.0,
        "ielts_breakdown": {"listening": 7, "reading": 7, "speaking": 6.5, "writing": 6.5},
        "sat_total": 1300,
        "intended_major": "Computer Science",
        "budget_usd_per_year": 30000,
        "preferred_countries": ["USA", "Canada"],
        "target_intake": "Fall 2026",
        "scholarship_required": True,
        "extracurriculars": "Debate Club",
        "achievements": "First place hackathon",
        "personal_statement_notes": "Focus on resilience."
    }
    
    print("Submitting Case to AI...")
    resp = requests.post(f"{BASE_URL}/cases", json=payload, headers=headers)
    if resp.status_code != 201:
        print(f"Failed to create case: {resp.text}")
        return
        
    case_id = resp.json()['data']['case_id']
    status = resp.json()['data']['status']
    print(f"Case Created: {case_id} | Init Status: {status}")
    
    # 3. Poll for Status Change
    timeout = 30 # seconds
    for i in range(timeout):
        c_resp = requests.get(f"{BASE_URL}/cases/{case_id}", headers=headers)
        if c_resp.status_code == 200:
            current_status = c_resp.json()['data']['status']
            if current_status in ['done', 'human_review', 'failed']:
                print(f"Wait Completed! Final Status: {current_status}")
                if current_status == 'done':
                    conf = c_resp.json()['data'].get('ai_confidence', 'N/A')
                    print(f"AI Matrix Processed. Confidence: {conf}%")
                return
            else:
                print(f"[{i}s] Status continues to reflect: {current_status}")
        time.sleep(1)
        
    print("AI Workflow Timed Out!")

run_ai_test()
