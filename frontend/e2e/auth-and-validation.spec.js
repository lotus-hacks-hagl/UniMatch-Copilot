import { test, expect } from '@playwright/test';
import { readRunState } from './helpers/state.js';
import { loginThroughUi, acceptNextDialog } from './helpers/ui.js';

test.describe('E2E Auth and validation', () => {
  test('@deterministic @live invalid login stays on auth view', async ({ page }) => {
    await page.goto('/auth');
    await page.getByTestId('auth-username').fill('invalid-user');
    await page.getByTestId('auth-password').fill('wrong-password');
    await page.getByTestId('auth-submit').click();

    await expect(page).toHaveURL(/\/auth$/);
    await expect(page.getByText(/invalid username or password|authentication failed/i)).toBeVisible();
  });

  test('@deterministic @live new case blocks submit without IELTS or SAT', async ({ page }) => {
    const state = await readRunState();

    await loginThroughUi(page, {
      username: state.admin.username,
      password: state.admin.password,
      expectedPath: '/admin/teachers',
    });

    await page.goto('/cases/new');
    await expect(page).toHaveURL(/\/cases\/new$/);

    await page.getByTestId('new-case-full-name').fill(`E2E Validation ${state.runId}`);
    await page.getByTestId('new-case-gpa-raw').fill('8.3');
    await page.getByTestId('new-case-continue').click();
    await page.getByTestId('new-case-major').selectOption('Computer Science');
    await page.getByRole('button', { name: 'USA' }).click();
    await page.getByTestId('new-case-budget').fill('32000');
    await page.getByTestId('new-case-continue').click();
    await page.getByTestId('new-case-intake').selectOption('Fall 2026');

    const dialogMessagePromise = acceptNextDialog(page);
    await page.getByTestId('new-case-submit').click();
    await expect(page).toHaveURL(/\/cases\/new$/);
    const dialogMessage = await dialogMessagePromise;
    expect(dialogMessage).toMatch(/please provide either ielts or sat score/i);
  });
});
