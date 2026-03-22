# Full System Integration Verification Plan (BE + FE)

## 1. Context & Objectives
To conduct a comprehensive, end-to-end integration audit of the UniMatch Copilot application. The goal is to guarantee 100% seamless communication between the Vue.js Frontend and the Backend API, ensuring all API calls, state management, and background tasks (like AI processing) execute flawlessly.

## 2. Agent Assignments
- **`qa-engineer`** (or **`orchestrator`**): To coordinate the testing phases.
- **`frontend-specialist`**: To monitor Vue DevTools, Pinia stores, and Vue Router guard interactions.
- **`backend-specialist`**: To audit server logs, check POST/GET payload parity, and verify JSON response formats.

## 3. Task Breakdown (Integration Audit)

### Phase 1: Authentication & Authorization Flow
- [ ] Verify Login/Register APIs (`/api/auth/login`, `/api/auth/register`).
- [ ] Validate JWT token storage and Axios interceptor injection.
- [ ] Test Role-Based Access Control (RBAC): Admin vs. Teacher route protection.
- [ ] Verify token expiration and 401 Unauthorized handling (graceful logout).

### Phase 2: CRUD Operations & State Management
- [ ] **Cases Module:** Verify payload construction for creating a new Case (`POST /cases`) and fetching case lists (`GET /cases` with filters).
- [ ] **Students Module:** Ensure student profile data binds correctly from FE forms to BE structs without dropping fields (e.g., GPA normalization, target_intake).
- [ ] **University KB:** Test the KB Sync background job trigger and pagination/filtering of the Universities endpoint.

### Phase 3: AI Engine & Background Tasks Execution
- [ ] **Case Processing:** Trigger the AI match analysis. Ensure the backend properly dispatches the workload to the AI Service (TinyFish).
- [ ] **Polling/Real-time Updates:** Verify that the frontend correctly polls or updates Case Status transitions (`pending_ai` -> `processing` -> `done`/`human_review`).
- [ ] **Global Dashboard KPIs:** Ensure frontend `stats` update reactively after backend processing.

### Phase 4: Edge Cases & Error Handling
- [ ] Simulate backend 500 errors and ensure the frontend displays friendly Toast/Alert notifications without freezing.
- [ ] Test empty data states and pagination edge-cases in the table views.
- [ ] Validate Cross-Origin Resource Sharing (CORS) configurations if applicable.

## 4. Verification Checklist (Testing Checkpoints)
- [ ] User can register, login, and persist session across page refresh.
- [ ] User can create a new case and see it transition through AI processing to "Done" automatically.
- [ ] Administrator can view the Teacher Management panel and update teacher verification statuses.
- [ ] Zero unhandled Promise rejections or console errors in the browser DevTools during a full user journey.
