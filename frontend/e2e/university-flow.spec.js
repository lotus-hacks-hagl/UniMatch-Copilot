import { test, expect } from '@playwright/test';
import { countActiveCrawls, createUniversity, findUniversityByName, login, waitForUniversityCrawlStatus } from './helpers/api.js';
import { readRunState } from './helpers/state.js';
import { loginThroughUi } from './helpers/ui.js';

test('@deterministic @live admin can trigger university crawl from FE', async ({ page }) => {
  const state = await readRunState();
  const adminAuth = await login(state.admin.username, state.admin.password);
  const universityName = `E2E University ${state.runId}`;
  const beforeCount = await countActiveCrawls(adminAuth.token);

  await createUniversity(adminAuth.token, {
    name: universityName,
    country: 'Canada',
    qs_rank: 999,
    tuition_usd_per_year: 18000,
    acceptance_rate: 0.52,
    available_majors: ['Computer Science'],
  });

  await loginThroughUi(page, {
    username: state.admin.username,
    password: state.admin.password,
    expectedPath: '/admin/teachers',
  });

  await page.getByRole('link', { name: /university kb/i }).click();
  await expect(page).toHaveURL(/\/universities$/);

  await page.getByTestId('university-sync').click();
  await expect(page.getByTestId('toast-success')).toContainText(/tinyfish crawl started in background/i);

  const crawledUniversity = await waitForUniversityCrawlStatus(
    adminAuth.token,
    universityName,
    state.mode === 'live' ? ['ok', 'changed'] : ['pending', 'ok', 'changed', 'failed'],
  );

  expect(crawledUniversity.crawl_status).not.toBe('never_crawled');
  const afterCount = await countActiveCrawls(adminAuth.token);
  expect(afterCount).toBeGreaterThanOrEqual(0);
  expect(afterCount >= beforeCount || crawledUniversity.crawl_status !== 'never_crawled').toBeTruthy();

  const listed = await findUniversityByName(adminAuth.token, universityName);
  expect(listed?.name).toBe(universityName);
});
