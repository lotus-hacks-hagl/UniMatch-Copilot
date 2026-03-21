import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/authStore'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/auth',
      name: 'auth',
      component: () => import('../views/AuthView.vue'),
      meta: { public: true }
    },
    {
      path: '/unverified',
      name: 'unverified',
      component: () => import('../views/UnverifiedView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/',
      redirect: '/cases'
    },
    {
      path: '/cases',
      name: 'cases',
      component: () => import('../views/CasesView.vue'),
      meta: { title: 'Case overview', sub: 'Saturday, 21 Mar 2026', requiresAuth: true, requiresVerification: true }
    },
    {
      path: '/cases/new',
      name: 'new-case',
      component: () => import('../views/NewCaseView.vue'),
      meta: { title: 'New Case', sub: 'Create a student profile', requiresAuth: true, requiresVerification: true }
    },
    {
      path: '/cases/:id',
      name: 'case-detail',
      component: () => import('../views/CaseDetailView.vue'),
      meta: { title: 'Case detail', requiresAuth: true, requiresVerification: true }
    },
    {
      path: '/students',
      name: 'students',
      component: () => import('../views/StudentsView.vue'),
      meta: { title: 'All students', requiresAuth: true, requiresVerification: true }
    },
    {
      path: '/admin/teachers',
      name: 'admin-teachers',
      component: () => import('../views/TeacherManagementView.vue'),
      meta: { title: 'Teacher Management', sub: 'Admin Control Panel', requiresAuth: true, requiresAdmin: true }
    },
    {
      path: '/universities',
      name: 'universities',
      component: () => import('../views/UniversityKBView.vue'),
      meta: { title: 'University KB', requiresAuth: true, requiresVerification: true }
    },
    {
      path: '/analytics',
      name: 'analytics',
      component: () => import('../views/AnalyticsView.vue'),
      meta: { title: 'Analytics', requiresAuth: true, requiresVerification: true }
    },
    {
      path: '/review-queue',
      name: 'review-queue',
      component: () => import('../views/ReviewQueueView.vue'),
      meta: { title: 'Review Queue', requiresAuth: true, requiresVerification: true }
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/'
    }
  ]
})

router.beforeEach((to, from, next) => {
  const auth = useAuthStore()
  
  // 1. Auth check
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return next('/auth')
  }

  // 2. Already logged in redirect
  if (to.name === 'auth' && auth.isAuthenticated) {
    return auth.isAdmin ? next('/admin/teachers') : next('/cases')
  }

  // 3. Admin strictly required
  if (to.meta.requiresAdmin && !auth.isAdmin) {
    return next('/cases')
  }

  // 4. Verification Check for teachers
  if (to.meta.requiresVerification && !auth.isAdmin && !auth.isVerified) {
    return next('/unverified')
  }

  next()
})

export default router
