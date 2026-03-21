import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/cases'
    },
    {
      path: '/cases',
      name: 'cases',
      component: () => import('../views/CasesView.vue'),
      meta: { title: 'Case overview', sub: 'Saturday, 21 Mar 2026 — 47 active cases' }
    },
    {
      path: '/cases/new',
      name: 'new-case',
      component: () => import('../views/NewCaseView.vue'),
      meta: { title: 'New Case', sub: 'Create a student profile' }
    },
    {
      path: '/cases/:id',
      name: 'case-detail',
      component: () => import('../views/CaseDetailView.vue'),
      meta: { title: 'Case detail', sub: '' }
    },
    {
      path: '/students',
      name: 'students',
      component: () => import('../views/StudentsView.vue'),
      meta: { title: 'All students', sub: 'Manage registered students' }
    },
    {
      path: '/queues',
      name: 'queues',
      component: () => import('../views/ReviewQueueView.vue'),
      meta: { title: 'Review queue', sub: '3 cases awaiting human review' }
    },
    {
      path: '/universities',
      name: 'universities',
      component: () => import('../views/UniversityKBView.vue'),
      meta: { title: 'University KB', sub: '284 universities automatically synced' }
    },
    {
      path: '/analytics',
      name: 'analytics',
      component: () => import('../views/AnalyticsView.vue'),
      meta: { title: 'Analytics', sub: 'Performance tracking & ROI' }
    }
  ]
})

export default router
