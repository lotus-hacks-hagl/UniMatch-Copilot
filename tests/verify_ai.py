import os
import time

import requests


BASE_URL = os.getenv("VERIFY_BASE_URL", "http://localhost:8894/api/v1")
USERNAME = os.getenv("VERIFY_USERNAME", "admin@unimatch.com")
PASSWORD = os.getenv("VERIFY_PASSWORD", "admin@123")
AI_SERVICE_BASE_URL = os.getenv("VERIFY_AI_SERVICE_BASE_URL", "http://localhost:8895")
VERIFY_DEBUG_JOBS = os.getenv("VERIFY_DEBUG_JOBS", "false").lower() == "true"


def ensure_success(resp, label):
    if resp.status_code >= 400:
        raise RuntimeError(f"{label} failed ({resp.status_code}): {resp.text}")


def run_ai_test():
    print("=== UniMatch AI End-to-End Verify ===")

    ai_health = requests.get(f"{AI_SERVICE_BASE_URL}/health", timeout=20)
    ensure_success(ai_health, "ai service health")
    print(f"AI service health: {ai_health.json()}")

    login = requests.post(
        f"{BASE_URL}/auth/login",
        json={"username": USERNAME, "password": PASSWORD},
        timeout=20,
    )
    ensure_success(login, "login")
    token = login.json()["data"]["token"]
    headers = {"Authorization": f"Bearer {token}"}
    print(f"Logged in as {USERNAME}")

    universities = requests.get(f"{BASE_URL}/universities?limit=5", headers=headers, timeout=20)
    ensure_success(universities, "list universities")
    total_universities = universities.json().get("meta", {}).get("total", 0)
    print(f"University KB count: {total_universities}")

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
        "personal_statement_notes": "Focus on resilience.",
    }

    print("Submitting case to AI service...")
    created = requests.post(f"{BASE_URL}/cases", json=payload, headers=headers, timeout=20)
    ensure_success(created, "create case")
    case_id = created.json()["data"]["case_id"]
    ai_job_id = None
    print(f"Case created: {case_id}")

    final_case = None
    for attempt in range(25):
        detail = requests.get(f"{BASE_URL}/cases/{case_id}", headers=headers, timeout=20)
        ensure_success(detail, "get case")
        case_data = detail.json()["data"]
        ai_job_id = case_data.get("ai_job_id")
        status = case_data["status"]
        print(f"[{attempt}] case status: {status}")
        if status in ["done", "human_review", "failed"]:
            final_case = case_data
            break
        time.sleep(1)

    if final_case is None:
        raise RuntimeError("case workflow timed out")

    print(f"Final status: {final_case['status']}")
    print(f"Recommendations: {len(final_case.get('recommendations') or [])}")
    print(f"AI confidence: {final_case.get('ai_confidence')}")
    if VERIFY_DEBUG_JOBS and ai_job_id:
        debug_resp = requests.get(f"{AI_SERVICE_BASE_URL}/jobs/{ai_job_id}", timeout=20)
        ensure_success(debug_resp, "get ai debug job")
        print(f"AI debug job: {debug_resp.json()['data']}")

    if final_case["status"] not in ["done", "human_review"]:
        raise RuntimeError("case did not reach reportable state")

    report = requests.post(f"{BASE_URL}/cases/{case_id}/report", headers=headers, timeout=20)
    ensure_success(report, "request report")
    print("Report request accepted")

    for attempt in range(15):
        detail = requests.get(f"{BASE_URL}/cases/{case_id}", headers=headers, timeout=20)
        ensure_success(detail, "get case after report")
        case_data = detail.json()["data"]
        if case_data.get("report_data"):
            print(f"Report ready on attempt {attempt}")
            print(f"Report summary: {case_data['report_data'].get('summary')}")
            break
        time.sleep(1)
    else:
        raise RuntimeError("report was not generated in time")

    crawl = requests.post(f"{BASE_URL}/universities/crawl-all", headers=headers, timeout=20)
    ensure_success(crawl, "crawl all universities")
    print(f"Crawl trigger response: {crawl.json()['data']}")

    active = requests.get(f"{BASE_URL}/universities/crawl-active", headers=headers, timeout=20)
    ensure_success(active, "count active crawls")
    print(f"Active crawls: {active.json()['data']['active_crawls']}")

    print("VERIFY_OK")


if __name__ == "__main__":
    run_ai_test()
