# PLAN-rbac-dashboard

## 🎯 Goal
Introduce a robust Role-Based Access Control (RBAC) system to UniMatch Copilot. This separates the generic "user" into `Super Admin` and `Teacher` (Counselor), adding proper verification and management features.

## 🧠 Brainstorm & Feature Gaps Identified
Based on current MVP implementation, the system lacks authentication and agency-scale management. We require:
1. **Auth Module:** Login/Register pages with JWT.
2. **Account Verification:** Teachers registering must be manually verified by the Super Admin before they gain access.
3. **Admin Dashboard (Teacher Management):** A specialized UI for the Super Admin to list out teachers, view their metrics (cases solved), and approve/suspend their accounts.
4. **My Cases (Teacher View):** Teachers should have a filtered view restricted to cases assigned to them, rather than a global feed pool.

---

## 🏗️ Implementation Strategy

### Phase 1: Backend (Golang) Foundation
- **Database Schema Updates:** 
  - Create table `users` (`id`, `email`, `password_hash`, `role` [admin/teacher], `is_verified`).
  - Update `cases` table to link `assigned_teacher_id` instead of a string `created_by`.
- **Auth APIs:** `POST /api/v1/auth/register`, `POST /api/v1/auth/login`.
- **Admin APIs:** `GET /api/v1/admin/teachers`, `PUT /api/v1/admin/teachers/:id/verify`.
- **Middleware:** JWT validation & RBAC (Admin-only routes).

### Phase 2: Frontend (Vue3) UI & Logic
- **Auth Views:** Create `LoginView.vue` and `RegisterView.vue`.
- **Pinia Auth Store:** Handle tokens, user roles, and interceptors.
- **Admin Dashboard View:** Create `TeacherManagementView.vue` (data table with Approve/Suspend buttons).
- **Navigation Guards:** Redirect unverified or unauthenticated users to Login; limit Adrmin routes to actual Admins.
- **Stitch UI Generation:** Utilize Stitch MCP loop to design breathtaking, premium-looking Auth and Admin Dashboard screens.

---

## 🛑 Socratic Gate Questions (For User Confirmation)

Before we instruct Stitch to design screens or write the backend code, please confirm:
1. **Business Logic on Cases:** Can a "Manager/Admin" unilaterally assign cases to a Teacher, or does the Teacher click a button to "Claim" a pending case from a global queue?
2. **First User Problem:** How is the *very first* Super Admin created? (A CLI script/DB seeder, or the first person to register becomes Admin automatically?)
3. **Design Aesthetics:** For the Auth and Admin screens driven by Stitch, do you prefer a clean minimalist vibe (like Vercel/Stripe) or a dynamic, colorful dashboard (Glassmorphism, vibrant gradients)?
