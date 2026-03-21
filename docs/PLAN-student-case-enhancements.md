# PLAN: Student & Case Enhancements

This plan outlines the steps to implement Student CRUD, Case background text, Re-analyze functionality, and various UI/UX improvements (Avatar fixes, Dashboard real-data integration).

## Proposed Changes

### 1. Backend Enhancement (Go)

#### [MODIFY] [student.go](file:///d:/Backend/UniMatch-Copilot/backend/internal/model/student.go)
- Add `DeletedAt gorm.DeletedAt` for soft delete support.
- Add `BackgroundText string` field.

#### [MODIFY] [activity_log.go](file:///d:/Backend/UniMatch-Copilot/backend/internal/model/activity_log.go)
- Add `EventCaseNote = "case_note"` constant.

#### [MODIFY] [students_dto.go](file:///d:/Backend/UniMatch-Copilot/backend/internal/dto/students_dto.go)
- Add `UpdateStudentRequest` DTO.

#### [MODIFY] [cases_dto.go](file:///d:/Backend/UniMatch-Copilot/backend/internal/dto/cases_dto.go)
- Add `BackgroundText` to `CreateCaseRequest`.
- Add `CaseNoteRequest` DTO.

#### [IMPLEMENT] Student CRUD & Re-analyze
- Update `StudentRepository` and `StudentService` with `Update`, `Delete` (soft), and `GetByID`.
- Update `CasesService` with `AddNote` and `ReAnalyze`.
- Update `DashboardRepository` to ensure queries are accurate and not returning placeholders.

---

### 2. AI Service Enhancement (ai-service-go)

#### [MODIFY] [models.go](file:///d:/Backend/UniMatch-Copilot/ai-service-go/internal/models/models.go) (or similar)
- Add `BackgroundText` to `AnalyzeInput`.

#### [MODIFY] [worker.go](file:///d:/Backend/UniMatch-Copilot/ai-service-go/internal/worker/worker.go) (or similar)
- Update prompt to include `Student Background`.

---

### 3. Frontend Enhancement (Vue)

#### [MODIFY] [api.js](file:///d:/Backend/UniMatch-Copilot/frontend/src/services/api.js)
- Add interceptor to auto-unwrap `res.data.data`.

#### [NEW] [avatar.js](file:///d:/Backend/UniMatch-Copilot/frontend/src/utils/avatar.js)
- Centralized helper for display name and avatar initials.

#### [MODIFY] [NewCaseView.vue](file:///d:/Backend/UniMatch-Copilot/frontend/src/views/NewCaseView.vue)
- Add "Student Background" textarea.

#### [MODIFY] [CaseDetailView.vue](file:///d:/Backend/UniMatch-Copilot/frontend/src/views/CaseDetailView.vue)
- Implement "Documents" tab (Report summary & download).
- Implement "Communication" tab (Activity log timeline + Add Note).

#### [NEW] [StudentDetailView.vue](file:///d:/Backend/UniMatch-Copilot/frontend/src/views/StudentDetailView.vue)
- View/Edit/Delete student details.

---

## Verification Plan

### Automated Tests
- `npm test` or `go test ./...`
- Playwright E2E: Create Case -> Analyze -> Add Note -> Re-Analyze -> Delete Student.

### Manual Verification
- Verify Dashboard stats change after creating/deleting cases.
- Verify Avatar fallback works for "Unnamed Student".
- Verify Background Text is sent to AI and reflected in the summary.
