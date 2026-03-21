export function getE2EConfig() {
  return {
    feBaseURL: process.env.E2E_FE_BASE_URL || 'http://127.0.0.1:5173',
    apiBaseURL: process.env.E2E_API_BASE_URL || 'http://127.0.0.1:8080/api/v1',
    aiBaseURL: process.env.E2E_AI_BASE_URL || 'http://127.0.0.1:9000',
    mode: process.env.E2E_MODE || 'deterministic',
    stack: process.env.E2E_STACK || 'auto',
    adminUsername: process.env.E2E_ADMIN_USERNAME || 'admin',
    adminPassword: process.env.E2E_ADMIN_PASSWORD || 'admin@123',
    teacherUsername: process.env.E2E_TEACHER_USERNAME || 'teacher.e2e@unimatch.com',
    teacherPassword: process.env.E2E_TEACHER_PASSWORD || 'teacher@123',
  };
}

export function requireLiveProviderKeys() {
  const missing = ['EXA_API_KEY', 'TINYFISH_API_KEY', 'OPENAI_API_KEY'].filter((key) => !process.env[key]);
  if (missing.length > 0) {
    throw new Error(`E2E live mode requires provider keys: ${missing.join(', ')}`);
  }
}
