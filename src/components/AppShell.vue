<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useQueueStore } from '../stores/queueStore'

const route = useRoute()
const router = useRouter()
const queueStore = useQueueStore()

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

const navItems = [
  { name: 'Cases', path: '/cases', iconColor: 'bg-primary' },
  { name: 'Students', path: '/students', iconColor: 'bg-gray-300' },
  { name: 'Review queue', path: '/queues', iconColor: 'bg-red-500', badge: computed(() => queueStore.pendingCount) },
  { name: 'University KB', path: '/universities', iconColor: 'bg-safe' },
  { name: 'Analytics', path: '/analytics', iconColor: 'bg-reach' }
]

const isActive = (path) => {
  if (path === '/cases' && route.path.startsWith('/cases')) return true
  return route.path === path
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
          <span class="w-[7px] h-[7px] rounded-full bg-gray-300"></span> Settings
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
            Export
          </button>
          <button 
            @click="router.push('/cases/new')"
            class="px-3.5 py-1.5 rounded-lg text-[13px] font-medium border border-primary bg-primary text-white hover:bg-primary-hover transition-colors whitespace-nowrap"
          >
            + New case
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
