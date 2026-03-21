<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useQueueStore } from '../stores/queueStore'
import { useI18n } from 'vue-i18n'

const route = useRoute()
const router = useRouter()
const queueStore = useQueueStore()
const { t, locale } = useI18n()

let syncInterval = null

onMounted(() => {
  queueStore.fetchPendingCount()
  queueStore.fetchSyncCount()
  syncInterval = setInterval(() => {
    queueStore.fetchSyncCount()
    queueStore.fetchPendingCount()
  }, 5000)
})

onUnmounted(() => {
  if (syncInterval) clearInterval(syncInterval)
})

const navItems = computed(() => [
  { name: t('common.cases'), path: '/cases', iconColor: 'bg-primary' },
  { name: t('common.students'), path: '/students', iconColor: 'bg-gray-300' },
  { name: t('common.reviews'), path: '/queues', iconColor: 'bg-red-500', badge: queueStore.pendingCount },
  { name: t('common.kb'), path: '/universities', iconColor: 'bg-safe' },
  { name: t('common.analytics'), path: '/analytics', iconColor: 'bg-reach' }
])

const isActive = (path) => {
  if (path === '/cases' && route.path.startsWith('/cases')) return true
  return route.path === path
}

const toggleLanguage = () => {
  locale.value = locale.value === 'en' ? 'vi' : 'en'
}
</script>

<template>
  <div class="flex h-screen overflow-hidden bg-bg text-text text-sm">
    <!-- Sidebar -->
    <aside class="w-[210px] shrink-0 bg-surface border-r border-black/10 flex flex-col">
      <div class="px-4 py-[18px] pb-3.5 text-base font-medium border-b border-black/10 shrink-0">
        Uni<span class="text-primary">Match</span> <span class="text-[11px] text-text-muted font-normal">Copilot</span>
      </div>
      
      <nav class="flex-1 px-2 py-2.5 flex flex-col gap-0.5 overflow-hidden">
        <router-link 
          v-for="item in navItems" 
          :key="item.path"
          :to="item.path"
          class="flex items-center gap-2.5 px-2.5 py-2 rounded-lg text-[13px] transition-colors shrink-0"
          :class="isActive(item.path) ? 'bg-secondary text-primary font-medium' : 'text-text-muted hover:bg-gray-50/50'"
        >
          <span class="w-[7px] h-[7px] rounded-full shrink-0" :class="item.iconColor"></span>
          {{ item.name }}
          <span 
            v-if="item.badge && item.badge > 0" 
            class="ml-auto bg-red-50 text-red-600 text-[10px] font-medium px-1.5 py-0.5 rounded-full"
          >
            {{ item.badge }}
          </span>
        </router-link>
      </nav>

      <div class="px-2 py-3 border-t border-black/10 shrink-0">
        <div class="flex items-center gap-2 px-2.5 py-2 text-[12px] text-text-muted cursor-pointer hover:bg-gray-50/50 rounded-lg mb-1">
          <svg class="w-3.5 h-3.5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path></svg>
          <span class="flex-1">{{ $t('common.settings') }}</span>
        </div>
        <div @click="toggleLanguage" class="flex items-center gap-2 px-2.5 py-2 text-[12px] text-text-muted cursor-pointer hover:bg-gray-50/50 rounded-lg mb-4">
          <svg class="w-3.5 h-3.5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5C11.783 10.77 8.07 15.61 3 18.129"></path></svg>
          <span class="flex-1">{{ $t('common.language') }}</span>
          <span class="text-[10px] font-bold uppercase bg-black/5 px-1.5 py-0.5 rounded">{{ locale }}</span>
        </div>
        <div class="flex items-center gap-2.5 px-2.5 py-2">
          <div class="w-8 h-8 rounded-full bg-secondary text-primary font-medium text-xs flex items-center justify-center shrink-0">
            TN
          </div>
          <div>
            <div class="text-xs font-medium text-text">Trang Nguyen</div>
            <div class="text-[11px] text-text-muted">Senior counselor</div>
          </div>
        </div>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="flex-1 flex flex-col overflow-hidden min-w-0">
      <!-- Topbar -->
      <header class="px-5 py-3 bg-surface border-b border-black/10 flex items-center justify-between shrink-0">
        <div>
          <h1 class="text-[15px] font-medium">{{ route.meta.title || 'UniMatch' }}</h1>
          <p class="text-[12px] text-text-muted mt-0.5">{{ route.meta.sub }}</p>
        </div>
        <div class="flex items-center gap-2.5">
          <div v-if="queueStore.syncCount > 0" class="inline-flex items-center gap-1.5 text-[11px] text-safe bg-safe/10 px-2.5 py-1 rounded-full border border-safe/40 whitespace-nowrap">
            <span class="w-1.5 h-1.5 rounded-full bg-safe animate-pulse"></span>
            TinyFish syncing {{ queueStore.syncCount }} universities
          </div>

          <button class="px-3.5 py-1.5 rounded-lg text-[13px] font-medium border border-black/15 hover:bg-gray-50 transition-colors whitespace-nowrap">
            {{ $t('common.export') }}
          </button>
          <button 
            @click="router.push('/cases/new')"
            class="px-3.5 py-1.5 rounded-lg text-[13px] font-medium border border-primary bg-primary text-white hover:bg-primary-hover transition-colors whitespace-nowrap"
          >
            {{ $t('common.newCase') }}
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
