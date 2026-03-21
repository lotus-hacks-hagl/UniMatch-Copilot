# E2E Test Suite and API Matrix Plan

## 1. Context & Objectives
To architect a bulletproof End-to-End (E2E) Test Suite that comprehensively covers the entire business workflow of the UniMatch Copilot application. Incorporating `webapp-testing` (Playwright) and `testing-patterns`, this plan outlines the strategy to validate:
- **UI E2E Flows**: Happy Paths (Login -> Create Case -> View Analytics) and multi-language toggles.
- **API Matrix Integration**: Deep validation of all backend routes (Auth, Cases, RBAC, University KB).
- **Edge Cases & Error Handling**: Invalid formatting, Unauthorized access routing, and Server timeout simulations.

## 2. Agent Assignments
- **`qa-engineer`**: To rig the Playwright architecture, define `playwright.config.ts`, and write browser-level user flow simulations.
- **`backend-specialist`**: To code the Data Seeding / Setup-Teardown hooks (wiping test databases before and after tests) ensuring 100% test idempotency.

## 3. Task Breakdown (Test Implementation)

### Phase 1: Test Infrastructure Setup
- [ ] Install Playwright (`npm init playwright@latest`) inside the frontend directory.
- [ ] Configure `playwright.config.ts` with multi-browser support (Chromium, Firefox, WebKit) and base URL mappings.
- [ ] Implement a `global-setup.ts` to programmatically purge DB test tables and cache Authentication states (Session Storage) to avoid redundant login delays.

### Phase 2: Core API Contract Validation (Headless Context)
- [ ] **Auth Gateway Route Testing**: Validate HTTP 201 (Valid Registration), HTTP 401 (Invalid Password), HTTP 400 (Duplicate User).
- [ ] **Cases & Dashboard API**: Mock and fetch `/cases` and `/dashboard/stats` validating JSON schema parity with expected DTOs.
- [ ] **AI Orchestration API**: Feed extreme payload constraints into `POST /cases` to trigger validation failures, verifying that Backend accurately routes HTTP 400 `VALIDATION_FAILED` parameters back to UI.

### Phase 3: Critical UI E2E Flows (Browser Render)
- [ ] **Flow 1 (The Happy Path)**: User logs in -> Navigates to 'New Case' -> Fills a multi-step form (IELTS 7.5, Comp Sci, $30k) -> Hits Submit -> Verifies the table correctly registers the status transition.
- [ ] **Flow 2 (Client Validation)**: Intentionally skip mandatory forms. Assert that correct Vuetify/Tailwind error states (red borders/text spans) visibly attach to inputs.
- [ ] **Flow 3 (RBAC Constraints)**: Assume a 'Teacher' (unverified) token -> Traverse directly to `/cases/new` -> System successfully redirects to the restricted `/unverified` guard page.
- [ ] **Flow 4 (i18n Translation Switch)**: Click the Language Toggle to 'VI' -> Assert that layout components dynamically update strings (e.g. "Case overview" becomes translated text).

## 4. Verification Checklist
- [ ] The command `npx playwright test` resolves 100% green without flakiness.
- [ ] Local Postgres Database maintains zero "dirty" test entries post-execution (Teardown hooks operational).
- [ ] Playwright HTML Report gracefully visualizes traces and screenshots of any intentionally forced failures.
