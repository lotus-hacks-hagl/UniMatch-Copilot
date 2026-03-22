<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useQueueStore } from '../stores/queueStore'
import { useAuthStore } from '../stores/authStore'
import { useI18n } from 'vue-i18n'

const route = useRoute()
const router = useRouter()
const queueStore = useQueueStore()
const authStore = useAuthStore()
const { t, locale } = useI18n()

let syncInterval = null

onMounted(() => {
  if (authStore.isAuthenticated && authStore.isVerified) {
    queueStore.fetchPendingCount()
    queueStore.fetchSyncCount()
    syncInterval = setInterval(() => {
      queueStore.fetchSyncCount()
      queueStore.fetchPendingCount()
    }, 5000)
  }
})

onUnmounted(() => {
  if (syncInterval) clearInterval(syncInterval)
})

const icons = {
  cases: `<svg class="w-[18px] h-[18px] shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M20.25 14.15v4.25c0 1.094-.787 2.036-1.872 2.18-2.087.277-4.216.42-6.378.42s-4.291-.143-6.378-.42c-1.085-.144-1.872-1.086-1.872-2.18v-4.25m16.5 0a2.18 2.18 0 00.75-1.661V8.706c0-1.081-.768-2.015-1.837-2.175a48.114 48.114 0 00-3.413-.387m4.5 8.006c-.194.165-.42.295-.673.38A23.978 23.978 0 0112 15.75c-2.648 0-5.195-.429-7.577-1.22a2.016 2.016 0 01-.673-.38m0 0A2.18 2.18 0 013 12.489V8.706c0-1.081.768-2.015 1.837-2.175a48.111 48.111 0 013.413-.387m7.5 0V5.25A2.25 2.25 0 0013.5 3h-3a2.25 2.25 0 00-2.25 2.25v.894m7.5 0a48.667 48.667 0 00-7.5 0M12 12.75h.008v.008H12v-.008z" /></svg>`,
  teachers: `<svg class="w-[18px] h-[18px] shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z" /></svg>`,
  students: `<svg class="w-[18px] h-[18px] shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z" /></svg>`,
  reviews: `<svg class="w-[18px] h-[18px] shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M15.182 15.182a4.5 4.5 0 01-6.364 0M21 12a9 9 0 11-18 0 9 9 0 0118 0zM9.75 9.75c0 .414-.168.75-.375.75S9 10.164 9 9.75 9.168 9 9.375 9s.375.336.375.75zm-.375 0h.008v.015h-.008V9.75zm5.625 0c0 .414-.168.75-.375.75s-.375-.336-.375-.75.168-.75.375-.75.375.336.375.75zm-.375 0h.008v.015h-.008V9.75z" /></svg>`,
  kb: `<svg class="w-[18px] h-[18px] shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6.042A8.967 8.967 0 006 3.75c-1.052 0-2.062.18-3 .512v14.25A8.987 8.987 0 016 18c2.305 0 4.408.867 6 2.292m0-14.25a8.966 8.966 0 016-2.292c1.052 0 2.062.18 3 .512v14.25A8.987 8.987 0 0018 18a8.967 8.967 0 00-6 2.292m0-14.25v14.25" /></svg>`,
  analytics: `<svg class="w-[18px] h-[18px] shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z" /></svg>`
}

const navItems = computed(() => {
  if (!authStore.isAuthenticated) return []

  const base = [
    { name: t('nav.cases'), path: '/cases', icon: icons.cases },
    { name: t('nav.students'), path: '/students', icon: icons.students },
    { name: t('nav.reviewQueue'), path: '/review-queue', icon: icons.reviews, badge: queueStore.pendingCount },
    { name: t('nav.universities'), path: '/universities', icon: icons.kb },
    { name: t('nav.analytics'), path: '/analytics', icon: icons.analytics }
  ]

  if (authStore.isAdmin) {
    base.splice(1, 0, { name: t('nav.adminTeachers'), path: '/admin/teachers', icon: icons.teachers })
  }

  return base
})

const pageTitle = computed(() => {
  if (!route.name) return route.meta.title || 'Case overview'
  const keyName = String(route.name).replace(/-([a-z])/g, (g) => g[1].toUpperCase())
  const key = `nav.${keyName}`
  return t(key) !== key ? t(key) : (route.meta.title || 'Overview')
})

const isActive = (path) => {
  if (path === '/cases' && route.path.startsWith('/cases')) return true
  return route.path === path
}

const toggleLanguage = () => {
  locale.value = locale.value === 'en' ? 'vi' : 'en'
}

const handleLogout = () => {
  authStore.logout()
  router.push('/auth')
}
</script>

<template>
  <div class="flex h-screen overflow-hidden bg-bg text-text text-sm">
    <!-- Sidebar -->
    <aside 
      v-if="authStore.isAuthenticated && (authStore.isAdmin || authStore.isVerified)"
      class="w-[240px] shrink-0 bg-white border-r border-black/5 flex flex-col pt-6 font-sans shadow-sm z-20"
    >
      <div class="px-6 pb-6 text-[15px] font-bold flex items-center gap-2.5">
        <svg viewBox="0 0 40 40" class="w-6 h-6" fill="none">
          <path d="M8 8 v14 c0 6.6 5.4 12 12 12 s12 -5.4 12 -12 v-14" stroke="#a32d2d" stroke-width="6" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
        <span class="text-[#18180f] tracking-tight">UniMatch Copilot</span>
      </div>
      
      <nav class="flex-1 py-2 flex flex-col gap-1.5 overflow-hidden">
        <router-link 
          v-for="item in navItems" 
          :key="item.path"
          :to="item.path"
          class="relative flex items-center gap-3 px-6 py-2.5 text-[14px] font-medium transition-all group pr-4"
          :class="isActive(item.path) ? 'text-[#a32d2d]' : 'text-[#6b6a62] hover:text-[#18180f] hover:bg-black/[0.02]'"
        >
          <!-- Active Left Bar Indicator -->
          <div v-if="isActive(item.path)" class="absolute left-0 top-0 bottom-0 w-[4px] bg-[#a32d2d] rounded-r-md"></div>
          
          <!-- Background Pill for Active State -->
          <div v-if="isActive(item.path)" class="absolute inset-y-0 left-2 right-4 bg-red-50 rounded-lg -z-10"></div>

          <div v-html="item.icon" :class="isActive(item.path) ? 'text-[#a32d2d]' : 'text-[#a8a79d] group-hover:text-[#6b6a62]'"></div>
          <span class="z-10">{{ item.name }}</span>
          
          <span 
            v-if="item.badge && item.badge > 0" 
            class="ml-auto bg-red-100 text-[#a32d2d] text-[11px] font-bold px-2 py-0.5 rounded-full z-10"
          >
            {{ item.badge }}
          </span>
        </router-link>
      </nav>

      <div class="p-4 flex flex-col gap-1 font-sans">
        <!-- Language -->
        <div @click="toggleLanguage" class="flex items-center gap-3 px-3 py-2 text-[14px] font-medium text-[#6b6a62] cursor-pointer hover:bg-black/[0.03] rounded-xl transition-colors">
          <svg class="w-[18px] h-[18px] text-[#a8a79d]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M12 21a9.004 9.004 0 008.716-6.747M12 21a9.004 9.004 0 01-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 017.843 4.582M12 3a8.997 8.997 0 00-7.843 4.582m15.686 0A11.953 11.953 0 0112 10.5c-2.974 0-5.768-.616-8.29-1.71m16.58 0a8.96 8.96 0 01.21 1.968c0 1.93-.36 3.774-1.01 5.452m-15.686 0a8.955 8.955 0 00-.21-1.968C2.511 12.43 2.15 10.587 2.15 8.657" /></svg>
          <span class="flex-1">{{ t('nav.language') }}</span>
          <span class="text-[11px] font-bold tracking-wider px-2 py-0.5 rounded-md bg-[#f4f5f7] border border-black/5 text-[#18180f] uppercase">{{ locale }}</span>
        </div>
        <!-- Sign Out -->
        <div @click="handleLogout" class="flex items-center gap-3 px-3 py-2 text-[14px] font-medium text-[#6b6a62] cursor-pointer hover:bg-black/[0.03] rounded-xl transition-colors mb-2">
          <svg class="w-[18px] h-[18px] text-[#a8a79d]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v13.5A2.25 2.25 0 007.5 21h6a2.25 2.25 0 002.25-2.25V15M12 9l-3 3m0 0l3 3m-3-3h12.75" /></svg>
          <span class="flex-1">{{ t('nav.signOut') }}</span>
        </div>
        <!-- Profile -->
        <div class="flex items-center gap-3 p-3 bg-white border border-black/5 hover:border-black/10 shadow-sm rounded-xl cursor-pointer transition-colors">
          <div class="w-9 h-9 rounded-full bg-[#f4f5f7] flex items-center justify-center -ml-1">
            <svg class="w-5 h-5 text-[#a8a79d]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z" /></svg>
          </div>
          <div class="overflow-hidden flex-1">
            <div class="text-[13px] font-bold text-[#18180f] truncate">{{ authStore.user?.username || 'admin1' }}</div>
            <div class="text-[11px] text-[#8a8980] uppercase tracking-wide font-medium mt-0.5">{{ authStore.user?.role || 'ADMIN' }}</div>
          </div>
          <svg class="w-4 h-4 text-[#a8a79d]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" /></svg>
        </div>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="flex-1 flex flex-col overflow-hidden min-w-0 bg-[#f7f8fa]">
      <!-- Topbar -->
      <header class="h-[88px] px-8 bg-[#f7f8fa] flex items-center justify-between shrink-0">
        <div>
          <h1 class="text-[24px] font-bold text-[#18180f] tracking-tight">{{ pageTitle }}</h1>
          <p class="text-[14px] text-[#6b6a62] mt-0.5">Saturday, 21 Mar 2026</p>
        </div>
        <div class="flex items-center gap-3">
          <div v-if="queueStore.syncCount > 0" class="inline-flex items-center gap-1.5 text-[11px] text-safe bg-safe/10 px-2.5 py-1 rounded-full border border-safe/40 whitespace-nowrap">
            <span class="w-1.5 h-1.5 rounded-full bg-safe animate-pulse"></span>
            {{ t('nav.syncing', { count: queueStore.syncCount }) }}
          </div>

          <button class="px-4 py-2 bg-white rounded-lg text-[13px] font-medium border border-black/10 hover:bg-gray-50 hover:shadow-sm hover:-translate-y-0.5 transition-all text-[#18180f] flex items-center gap-2">
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5M16.5 12L12 16.5m0 0L7.5 12m4.5 4.5V3" /></svg>
            {{ t('nav.export') }}
          </button>
          <button 
            @click="router.push('/cases/new')"
            class="px-4 py-2 rounded-lg text-[13px] font-bold border border-[#a32d2d] bg-[#a32d2d] text-white hover:bg-[#8B0000] hover:shadow-md hover:-translate-y-0.5 transition-all flex items-center gap-2"
          >
            {{ t('nav.newCase') }}
          </button>
        </div>
      </header>

      <!-- Scrollable body -->
      <div class="flex-1 overflow-y-auto w-full">
        <slot />
      </div>
    </main>
  </div>
</template>
