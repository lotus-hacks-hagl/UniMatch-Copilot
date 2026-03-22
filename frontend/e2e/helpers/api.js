import axios from 'axios';
import { getE2EConfig } from './config.js';

function createClient(baseURL) {
  return axios.create({
    baseURL,
    timeout: 30000,
    headers: {
      'Content-Type': 'application/json',
    },
  });
}

export async function probeHealth(url) {
  const response = await axios.get(url, { timeout: 10000 });
  return response.data;
}

export async function login(username, password) {
  const cfg = getE2EConfig();
  const client = createClient(cfg.apiBaseURL);
  const response = await client.post('/auth/login', { username, password });
  return response.data.data;
}

export async function register(username, password) {
  const cfg = getE2EConfig();
  const client = createClient(cfg.apiBaseURL);
  const response = await client.post('/auth/register', { username, password });
  return response.data.data;
}

export function createAuthedClient(token, baseURL = getE2EConfig().apiBaseURL) {
  return axios.create({
    baseURL,
    timeout: 30000,
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
  });
}

export async function ensureAdminAccount(username, password) {
  try {
    return await login(username, password);
  } catch (error) {
    await register(username, password);
    return login(username, password);
  }
}

export async function ensureTeacherAccount(adminToken, username, password) {
  let effectiveUsername = username;
  let teacherLogin = null;
  try {
    teacherLogin = await login(effectiveUsername, password);
  } catch (error) {
    try {
      await register(effectiveUsername, password);
    } catch (registerError) {
      if (registerError.response?.status === 409 || registerError.response?.status === 500) {
        effectiveUsername = `${username}.${Date.now()}`;
        await register(effectiveUsername, password);
      } else {
        throw registerError;
      }
    }
    teacherLogin = await login(effectiveUsername, password);
  }

  const adminClient = createAuthedClient(adminToken);
  const teachersResponse = await adminClient.get('/admin/teachers');
  const teacher = (teachersResponse.data.data || []).find((item) => item.username === effectiveUsername);
  if (!teacher) {
    throw new Error(`Unable to locate teacher account ${effectiveUsername} after registration`);
  }
  if (!teacher.is_verified) {
    await adminClient.put(`/admin/teachers/${teacher.id}/verify`, { is_verified: true });
    teacherLogin = await login(effectiveUsername, password);
  }

  return {
    teacher,
    username: effectiveUsername,
    auth: teacherLogin,
  };
}

export function makeCasePayload(runId, label, overrides = {}) {
  return {
    full_name: `E2E ${label} ${runId}`,
    gpa_normalized: 3.74,
    gpa_raw: 8.5,
    gpa_scale: 10,
    ielts_overall: 7,
    sat_total: 1310,
    intended_major: 'Computer Science',
    budget_usd_per_year: 32000,
    preferred_countries: ['USA', 'Canada'],
    target_intake: 'Fall 2026',
    scholarship_required: true,
    extracurriculars: 'Debate Club, Robotics Team',
    achievements: 'Won a regional hackathon',
    personal_statement_notes: 'Prefers project-based learning',
    ...overrides,
  };
}

export async function createCase(adminToken, payload) {
  const client = createAuthedClient(adminToken);
  const response = await client.post('/cases', payload);
  return response.data.data;
}

export async function getCase(token, caseId) {
  const client = createAuthedClient(token);
  const response = await client.get(`/cases/${caseId}`);
  return response.data.data;
}

export async function listCases(token, params = {}) {
  const client = createAuthedClient(token);
  const response = await client.get('/cases', { params });
  return response.data.data || [];
}

export async function waitForCaseFinal(token, caseId, options = {}) {
  const timeoutMs = options.timeoutMs || 120000;
  const intervalMs = options.intervalMs || 2000;
  const startedAt = Date.now();

  while (Date.now() - startedAt < timeoutMs) {
    const current = await getCase(token, caseId);
    if (['done', 'human_review', 'failed'].includes(current.status)) {
      return current;
    }
    await new Promise((resolve) => setTimeout(resolve, intervalMs));
  }

  throw new Error(`Timed out waiting for case ${caseId} to leave processing state`);
}

export async function waitForReport(token, caseId, options = {}) {
  const timeoutMs = options.timeoutMs || 90000;
  const intervalMs = options.intervalMs || 2000;
  const startedAt = Date.now();

  while (Date.now() - startedAt < timeoutMs) {
    const current = await getCase(token, caseId);
    if (current.report_data) {
      return current;
    }
    await new Promise((resolve) => setTimeout(resolve, intervalMs));
  }

  throw new Error(`Timed out waiting for report data on case ${caseId}`);
}

export async function requestReport(token, caseId) {
  const client = createAuthedClient(token);
  const response = await client.post(`/cases/${caseId}/report`);
  return response.data.data;
}

export async function countActiveCrawls(token) {
  const client = createAuthedClient(token);
  const response = await client.get('/universities/crawl-active');
  return response.data.data.active_crawls;
}

export async function createUniversity(token, payload) {
  const client = createAuthedClient(token);
  const response = await client.post('/universities', payload);
  return response.data.data;
}

export async function findUniversityByName(token, name) {
  const client = createAuthedClient(token);
  const response = await client.get('/universities', {
    params: { search: name, page: 1, limit: 20 },
  });
  return (response.data.data || []).find((item) => item.name === name) || null;
}

export async function waitForUniversityCrawlStatus(token, name, acceptedStatuses, options = {}) {
  const timeoutMs = options.timeoutMs || 120000;
  const intervalMs = options.intervalMs || 3000;
  const startedAt = Date.now();

  while (Date.now() - startedAt < timeoutMs) {
    const university = await findUniversityByName(token, name);
    if (university && acceptedStatuses.includes(university.crawl_status)) {
      return university;
    }
    await new Promise((resolve) => setTimeout(resolve, intervalMs));
  }

  throw new Error(`Timed out waiting for university ${name} to reach statuses: ${acceptedStatuses.join(', ')}`);
}

export async function getJobDebug(jobId) {
  const cfg = getE2EConfig();
  const response = await axios.get(`${cfg.aiBaseURL}/jobs/${jobId}`, { timeout: 15000 });
  return response.data.data;
}
