# PLAN-verify-features.md - Full System Verification

This document outlines the step-by-step verification plan for the UniMatch Copilot system, focusing on the newly implemented Refined Auth, RBAC, and Case Management features.

---

## Phase 1: Authentication & Role Assignment

### Test Case 1.1: Registration Refinements
1. **Navigate** to `/auth` on a fresh session.
2. **Switch** to "Join UniMatch" (Sign Up).
3. **Verify** that the "Confirm Password" field is visible.
4. **Action:** Enter a username, a 5-character password, and a mismatched confirm password.
5. **Expect:** Validation errors for short password and mismatch.
6. **Action:** Enter valid data (e.g., `admin1`, `password123`).
7. **Expect:** Successful redirect.

### Test Case 1.2: First User is Super Admin
1. **Clear Database** (optional but recommended for this test).
2. **Register** the very first user (e.g., `admin_main`).
3. **Verify** the user is redirected to `/admin/teachers`.
4. **Verify** the side menu contains "Teacher Management".

### Test Case 1.3: Subsequent Users are Teachers (Unverified)
1. **Logout** as admin.
2. **Register** a second user (e.g., `teacher_beta`).
3. **Verify** the user is redirected to `/unverified`.
4. **Expect:** A "Pending Verification" message; no access to other menus.

---

## Phase 2: RBAC & Teacher Management

### Test Case 2.1: Admin Verifies Teacher
1. **Login** as the first `admin_main`.
2. **Navigate** to "Teacher Management".
3. **Verify** that `teacher_beta` appears in the list with status "Pending".
4. **Action:** Click "Verify" button.
5. **Expect:** Status changes to "Verified".

### Test Case 2.2: Verified Teacher Access
1. **Logout** as admin.
2. **Login** as `teacher_beta`.
3. **Expect:** Redirected to `/cases`.
4. **Verify:** Side menu shows "Cases", "Students", "Analytics" but NOT "Teacher Management".

---

## Phase 3: Case Management Workflow

### Test Case 3.1: New Case Submission (The "Fix" Verification)
1. **Navigate** to `/cases/new`.
2. **Step 1:** Enter "John Doe", GPA Raw: `8.5`, Scale: `/ 10`.
3. **Step 2:** Select "Computer Science", USA, Budget: `0` (Testing the 0-fix).
4. **Step 3:** Select "Fall 2026" and Submit.
5. **Expect:** No validation errors; redirected to Case Detail page.
6. **Network Check:** Verify `gpa_normalized: 3.4` was sent in the payload.

### Test Case 3.2: Case Pool & Claiming
1. **Logout** as `teacher_beta`.
2. **Login** as a different verified teacher (e.g., `teacher_gamma`).
3. **Navigate** to "Cases".
4. **Action:** Filter by "Unassigned" (if available) or look for a "Claim Case" button.
5. **Expect:** Ability to claim the case created by John Doe.
6. **Verify:** Once claimed, the case moves to "My Cases".

---

## Phase 4: AI Analysis & Report Editing

### Test Case 4.1: Profile Review
1. **Navigate** to a case in "Done" status (or mock the AI callback).
2. **Verify** "AI Analysis" tab shows Safe/Match/Reach universities.
3. **Verify** correct color coding (Green/Amber/Red).

### Test Case 4.2: Report Editor
1. **Navigate** to "Report Editor" tab.
2. **Action:** Modify the AI-generated text.
3. **Action:** Click "Save Changes".
4. **Action:** Click "Export PDF".
5. **Expect:** Loading state "Generating PDF..." appears.

---

## Phase 5: UI & Localization

### Test Case 5.1: Language Switching
1. **Toggle** language at the bottom left (EN/VI).
2. **Verify** that labels in the Auth form and Dashboard menus translate correctly.
3. **Verify** the "Confirm Password" label works in both languages.

---

## Success Criteria
- [ ] 0 Auth bypasses for unverified teachers.
- [ ] Admin unique access to Teacher Management.
- [ ] GPA Normalization accuracy verified in DB.
- [ ] All mandatory fields in `CreateCaseRequest` correctly handle null/zero.
