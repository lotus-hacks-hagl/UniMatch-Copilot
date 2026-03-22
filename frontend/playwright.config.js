import { defineConfig, devices } from '@playwright/test';

const feBaseURL = process.env.E2E_FE_BASE_URL || 'http://127.0.0.1:5173';
const apiBaseURL = process.env.E2E_API_BASE_URL || 'http://127.0.0.1:8894/api/v1';
const url = new URL(feBaseURL);
const isLocalFe = ['127.0.0.1', 'localhost'].includes(url.hostname);
const fePort = Number(url.port || 5173);

export default defineConfig({
  testDir: './e2e',
  timeout: 90 * 1000,
  expect: {
    timeout: 15 * 1000
  },
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: [['list'], ['html', { open: 'never' }]],
  globalSetup: './e2e/global-setup.js',
  use: {
    baseURL: feBaseURL,
    trace: 'on-first-retry',
    viewport: { width: 1440, height: 900 },
    video: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    }
  ],
  webServer: isLocalFe ? {
    command: `npm run dev -- --host 127.0.0.1 --port ${fePort}`,
    port: fePort,
    reuseExistingServer: false,
    env: {
      ...process.env,
      VITE_API_BASE_URL: apiBaseURL,
    },
  } : undefined,
});
