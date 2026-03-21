import { randomUUID } from 'crypto';
import { ensureAdminAccount, ensureTeacherAccount, probeHealth } from './helpers/api.js';
import { getE2EConfig, requireLiveProviderKeys } from './helpers/config.js';
import { writeRunState } from './helpers/state.js';

async function requireReachable(label, url) {
  try {
    return await probeHealth(url);
  } catch (error) {
    throw new Error(`${label} health check failed for ${url}: ${error.message}`);
  }
}

export default async function globalSetup() {
  const cfg = getE2EConfig();
  const runId = randomUUID().slice(0, 8);

  console.log(`[E2E] Starting ${cfg.mode} suite on stack=${cfg.stack}, run=${runId}`);

  if (cfg.mode === 'live') {
    requireLiveProviderKeys();
  }

  await requireReachable('Backend', `${cfg.apiBaseURL.replace(/\/api\/v1$/, '')}/health`);
  await requireReachable('AI service', `${cfg.aiBaseURL}/health`);

  const adminAuth = await ensureAdminAccount(cfg.adminUsername, cfg.adminPassword);
  const teacherState = await ensureTeacherAccount(adminAuth.token, cfg.teacherUsername, cfg.teacherPassword);

  await writeRunState({
    runId,
    mode: cfg.mode,
    stack: cfg.stack,
    feBaseURL: cfg.feBaseURL,
    apiBaseURL: cfg.apiBaseURL,
    aiBaseURL: cfg.aiBaseURL,
    admin: {
      username: cfg.adminUsername,
      password: cfg.adminPassword,
    },
    teacher: {
      username: teacherState.username,
      password: cfg.teacherPassword,
      id: teacherState.teacher.id,
    },
  });

  console.log(`[E2E] Prepared admin=${cfg.adminUsername} teacher=${teacherState.username}`);
}
