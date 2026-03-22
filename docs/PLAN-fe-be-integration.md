# Frontend & Backend API Integration Plan

## 1. Overview & Success Criteria
- **Goal:** Review the entire Frontend interface, replace all mocked API calls with real backend integrations, and create a sophisticated backend data initialization script to populate realistic, high-quality sample data.
- **Project Type:** FULLSTACK (involves both `frontend-specialist` and `backend-specialist` or `orchestrator`).
- **Success Criteria:**
  - 100% of frontend data is fetched live from the backend (no hardcoded/js mocks).
  - The database initialization script runs successfully and populates realistic data (Universities, Students, Cases) out of the box.
  - UI seamlessly handles real data variants (long text, empty lists, paginated data).

## 2. Tech Stack Context
- **Frontend:** Vue 3 + Tailwind CSS + Vite (plus Pinia & Vue Router)
- **Backend:** Node.js/Python (dependent on existing BE architecture) 
- **Database Initializer:** Python with `faker` or Node.js equivalent to generate the mock dataset.

## 3. Task Breakdown

### Phase 1: Context & Endpoint Auditing
- **Task 1.1:** Map out all existing frontend mock integrations (e.g. `src/services` or `src/api`).
- **Task 1.2:** Audit backend APIs to ensure endpoints exist for all mapped FE requests. Identify gaps.
  - *Agent:* `orchestrator`

### Phase 2: Beautiful Fake Data Initialization Script
- **Task 2.1:** Create a robust generation script (`init_data.py` or `seed.js`) targeting the backend database.
- **Task 2.2:** Generate highly realistic dummy data for:
  - Users / Accounts (Admin, Counselors, Students)
  - 30+ Universities with rich details (programs, tuition, stats, logos)
  - 50+ diverse Student profiles and Application Cases
- **Task 2.3:** Run script and verify DB population.
  - *Agent:* `backend-specialist`
  - *INPUT:* Database Schema -> *OUTPUT:* Populated DB -> *VERIFY:* Script runs without error & data looks realistic.

### Phase 3: Total Frontend API Integration
- **Task 3.1:** Rewire Axios/fetch calls in the Vue modules to hit real API endpoints. Ensure Auth tokens are passed in headers.
- **Task 3.2:** Handle Network States: Loading skeletons, Error boundaries (404/500/401 handling).
- **Task 3.3:** Clear out all old Vue mock services/static JSON.
  - *Agent:* `frontend-specialist`
  - *INPUT:* BE Data -> *OUTPUT:* FE State -> *VERIFY:* Vue DevTools shows data correctly loaded from BE.

### Phase 4: UI Review & Polish
- **Task 4.1:** Review Dashboard, University KB, Analytics, and Student views against the new heavy/realistic data.
- **Task 4.2:** Fix style breakages: Text overflow, pagination logic, scrolling behavior, missing avatars.
  - *Agent:* `frontend-specialist`
  - *INPUT:* Populated FE -> *OUTPUT:* Fixed Layout -> *VERIFY:* No UI clipping or console logs/errors.

## 4. Phase X: Final Verification
- [ ] Run linter `npm run lint`.
- [ ] Script executions: `python .agent/skills/performance-profiling/scripts/lighthouse_audit.py`
- [ ] Manual test: Walk through login, dashboard, case detail, and university list.
- [ ] Socratic Gate was respected!
