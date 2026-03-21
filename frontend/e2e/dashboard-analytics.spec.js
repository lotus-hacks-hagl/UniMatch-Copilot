import { test, expect } from '@playwright/test';
import { login, createAuthedClient } from './helpers/api.js';
import { readRunState } from './helpers/state.js';
import { loginThroughUi } from './helpers/ui.js';

test('@deterministic @live admin dashboard renders analytics from backend', async ({ page }) => {
  const state = await readRunState();

  await loginThroughUi(page, {
    username: state.admin.username,
    password: state.admin.password,
    expectedPath: '/admin/teachers',
  });

  const auth = await login(state.admin.username, state.admin.password);
  const client = createAuthedClient(auth.token);
  const statsResponse = await client.get('/dashboard/stats');
  const casesByDayResponse = await client.get('/dashboard/cases-by-day');

  const stats = statsResponse.data.data;
  const casesByDay = casesByDayResponse.data.data || [];

  await page.getByRole('link', { name: /case overview/i }).click();
  await expect(page).toHaveURL(/\/cases$/);

  await expect(page.getByTestId('cases-stats-cases-today')).toContainText(String(stats.casesToday));
  await expect(page.getByTestId('cases-stats-avg-processing')).toContainText(`${Math.round(stats.avgProcessingTime)}m`);
  await expect(page.getByTestId('cases-stats-awaiting-review')).toContainText(String(stats.awaitingReview));
  await expect(page.getByTestId('cases-stats-ai-confidence')).toContainText(`${Math.round((stats.aiConfidenceAvg || 0) * 100)}%`);

  await expect(page.getByTestId('cases-chart-cases-per-day').locator('canvas')).toBeVisible();
  await expect(page.getByTestId('cases-chart-match-tier').locator('canvas')).toBeVisible();
  await expect(page.getByTestId('cases-chart-escalation-trend').locator('canvas')).toBeVisible();

  if (casesByDay.length > 0) {
    await expect(page.getByText(/cases per day/i)).toBeVisible();
  }
});
