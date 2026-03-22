import { test, expect } from '@playwright/test';
import { createCase, getCase, login, makeCasePayload, waitForCaseFinal, waitForReport } from './helpers/api.js';
import { readRunState } from './helpers/state.js';
import { acceptNextDialog, loginThroughUi } from './helpers/ui.js';

test('@deterministic @live teacher can claim a case, edit summary, and generate report', async ({ page }) => {
  const state = await readRunState();
  const adminAuth = await login(state.admin.username, state.admin.password);
  const teacherAuth = await login(state.teacher.username, state.teacher.password);
  const casePayload = makeCasePayload(state.runId, 'Teacher Flow');
  const created = await createCase(adminAuth.token, casePayload);
  const finalCase = await waitForCaseFinal(adminAuth.token, created.case_id);

  await loginThroughUi(page, {
    username: state.teacher.username,
    password: state.teacher.password,
    expectedPath: '/cases',
  });

  await expect(page.getByText(casePayload.full_name)).toBeVisible({ timeout: 30000 });
  await page.getByText(casePayload.full_name).click();
  await expect(page).toHaveURL(new RegExp(`/cases/${created.case_id}$`));

  const claimDialog = acceptNextDialog(page);
  await page.getByTestId('case-claim').click();
  await expect(await claimDialog).toMatch(/case claimed successfully/i);
  await expect(page.getByTestId('case-tab-reportEditor')).toBeVisible();

  await page.getByTestId('case-tab-reportEditor').click();
  const updatedSummary = `Teacher refined summary for ${state.runId}`;
  await page.getByTestId('case-report-summary').fill(updatedSummary);

  const saveDialog = acceptNextDialog(page);
  await page.getByTestId('case-save-report-summary').click();
  await expect(await saveDialog).toMatch(/summary updated/i);

  const afterSave = await getCase(teacherAuth.token, created.case_id);
  expect(afterSave.profile_summary?.main_opinion || '').toContain(updatedSummary);

  const reportDialog = acceptNextDialog(page);
  await page.getByTestId('case-generate-report').click();
  await expect(await reportDialog).toMatch(/report generation triggered successfully/i);

  const reportedCase = await waitForReport(teacherAuth.token, created.case_id);
  expect(reportedCase.report_data?.summary || '').toContain(casePayload.full_name);

  await page.reload();
  await expect(page.getByRole('heading', { name: casePayload.full_name })).toBeVisible();
  await expect(page.getByTestId('case-tab-reportEditor')).toBeVisible();
  expect(['done', 'human_review']).toContain(finalCase.status);
});
