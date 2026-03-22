import { expect } from '@playwright/test';

export async function loginThroughUi(page, { username, password, expectedPath }) {
  await page.goto('/auth');
  await page.getByTestId('auth-username').fill(username);
  await page.getByTestId('auth-password').fill(password);
  await page.getByTestId('auth-submit').click();
  await expect(page).toHaveURL(new RegExp(expectedPath));
}

export async function logoutThroughUi(page) {
  await page.getByText('Sign Out').click();
  await expect(page).toHaveURL(/\/auth$/);
}

export async function acceptNextDialog(page) {
  return new Promise((resolve) => {
    page.once('dialog', async (dialog) => {
      const message = dialog.message();
      await dialog.accept();
      resolve(message);
    });
  });
}

export async function createCaseThroughUi(page, fullName) {
  await page.getByRole('button', { name: /new case/i }).click();
  await expect(page).toHaveURL(/\/cases\/new$/);

  await page.getByTestId('new-case-full-name').fill(fullName);
  await page.getByTestId('new-case-gpa-raw').fill('3.8');
  await page.getByTestId('new-case-ielts').fill('7');
  await page.getByTestId('new-case-continue').click();

  await page.getByTestId('new-case-major').selectOption('Computer Science');
  await page.getByRole('button', { name: 'USA' }).click();
  await page.getByRole('button', { name: 'Canada' }).click();
  await page.getByTestId('new-case-budget').fill('32000');
  await page.getByTestId('new-case-continue').click();

  await page.getByTestId('new-case-intake').selectOption('Fall 2026');
  await page.getByTestId('new-case-submit').click();
  await expect(page).toHaveURL(/\/cases\/[0-9a-f-]+$/);

  const segments = page.url().split('/');
  return segments[segments.length - 1];
}
