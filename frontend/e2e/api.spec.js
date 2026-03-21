import { test, expect } from '@playwright/test';

test.describe('Core API Contract Validation (Headless)', () => {
  const BASE_URL = 'http://localhost:8080/api/v1';

  test('Auth Gateway Route Testing - Should handle Registration and Rejection', async ({ request }) => {
    // Attempting to register the global e2e_admin account setup by the global-setup script again
    const dupRes = await request.post(`${BASE_URL}/auth/register`, {
      data: { username: 'e2e_admin', password: 'testpassword123' }
    });
    
    // Status 400 is expected because it fails on duplicate admin existence
    expect(dupRes.status()).toBe(400);

    // Invalid Password Login Check
    const failLogin = await request.post(`${BASE_URL}/auth/login`, {
      data: { username: 'e2e_admin', password: 'wrongpassword' }
    });
    expect(failLogin.status()).toBe(401);
    
    // Valid Login Check
    const successLogin = await request.post(`${BASE_URL}/auth/login`, {
      data: { username: 'e2e_admin', password: 'testpassword123' }
    });
    expect(successLogin.status()).toBe(200);
    const body = await successLogin.json();
    expect(body.data.token).toBeDefined();
  });

  test('Cases & Dashboard API Schema Mapping', async ({ request }) => {
    // Grab Token
    const loginRes = await request.post(`${BASE_URL}/auth/login`, {
      data: { username: 'e2e_admin', password: 'testpassword123' }
    });
    const { token } = (await loginRes.json()).data;
    
    const dashboardRes = await request.get(`${BASE_URL}/dashboard/stats`, {
      headers: { Authorization: `Bearer ${token}` }
    });
    expect(dashboardRes.status()).toBe(200);
    const dashBody = await dashboardRes.json();
    expect(dashBody.data).toHaveProperty('casesToday');
    expect(dashBody.data).toHaveProperty('awaitingReview');
  });

  test('AI Orchestration API Constraint Mapping (Budget Error)', async ({ request }) => {
    const loginRes = await request.post(`${BASE_URL}/auth/login`, {
      data: { username: 'e2e_admin', password: 'testpassword123' }
    });
    const { token } = (await loginRes.json()).data;
    
    // Intentionally dropping mandatory BudgetUsdPerYear to assert Server 400 routing back
    const payload = {
        full_name: "Constraint Test Subject",
        gpa_normalized: 3.75,
        gpa_raw: 8.5,
        gpa_scale: 10.0,
        ielts_overall: 7.0,
        sat_total: 1300,
        intended_major: "Computer Science"
        // BudgetUsdPerYear omitted! Expected backend to trigger min=0 valid check or req check
    };
    
    const failCaseRes = await request.post(`${BASE_URL}/cases`, {
      headers: { Authorization: `Bearer ${token}` },
      data: payload
    });
    
    expect(failCaseRes.status()).toBe(400);
    const responseData = await failCaseRes.json();
    expect(responseData.success).toBe(false);
    expect(responseData.error.code).toBe("VALIDATION_FAILED");
  });
});
