import { test, expect } from '@playwright/test';
import { getCase, getJobDebug, login, waitForCaseFinal } from './helpers/api.js';
import { readRunState } from './helpers/state.js';
import { createCaseThroughUi, loginThroughUi } from './helpers/ui.js';

test('@deterministic @live admin can create a case and see AI results end-to-end', async ({ page }) => {
  const state = await readRunState();

  await loginThroughUi(page, {
    username: state.admin.username,
    password: state.admin.password,
    expectedPath: '/admin/teachers',
  });

  await page.getByRole('link', { name: /case overview/i }).click();
  await expect(page).toHaveURL(/\/cases$/);

  const fullName = `E2E Admin Flow ${state.runId}`;
  const caseId = await createCaseThroughUi(page, fullName);

  const adminAuth = await login(state.admin.username, state.admin.password);
  const finalCase = await waitForCaseFinal(adminAuth.token, caseId);
  expect(['done', 'human_review']).toContain(finalCase.status);
  expect(finalCase.recommendations?.length || 0).toBeGreaterThan(0);

  await page.reload();
  await expect(page.getByText(/ai match verdict/i)).toBeVisible();
  await expect(page.getByRole('heading', { name: fullName })).toBeVisible();

  await page.getByTestId('case-tab-aiAnalysis').click();
  await expect(page.getByRole('heading', { name: finalCase.recommendations[0].university_name })).toBeVisible();
  await expect(page.getByText(/likelihood/i).first()).toBeVisible();

  if (state.mode === 'deterministic') {
    const refreshedCase = await getCase(adminAuth.token, caseId);
    if (refreshedCase.ai_job_id) {
      const debugJob = await getJobDebug(refreshedCase.ai_job_id);
      expect(debugJob.callback_status).toBe('delivered');
      expect(debugJob.search_attempts.length).toBe(5);
    }
  }
});
