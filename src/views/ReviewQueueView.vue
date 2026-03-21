<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../services/api'

const router = useRouter()
const queue = ref([])
const loading = ref(true)

const fetchQueue = async () => {
  try {
    const response = await api.get('/cases', {
      params: { status: 'human_review', page: 1, limit: 100 }
    })
    queue.value = response.data.cases || []
  } catch (error) {
    console.error('Failed to fetch human review queue', error)
  } finally {
    loading.value = false
  }
}

onMounted(fetchQueue)

const getAvatar = (name) => {
  if (!name) return '?'
  return name.substring(0, 2).toUpperCase()
}

const getSlaInfo = (createdAt) => {
  const createdDate = new Date(createdAt)
  const deadlineMs = createdDate.getTime() + 8 * 3600 * 1000
  const leftMs = deadlineMs - Date.now()
  
  if (leftMs <= 0) return { text: 'Overdue', class: 'text-reach bg-reach/10 border-reach/20' }
  const leftHrs = Math.floor(leftMs / 3600000)
  const leftMins = Math.floor((leftMs % 3600000) / 60000)
  
  const text = `${leftHrs}h ${leftMins}m left`
  let css = 'text-safe bg-safe/10 border-safe/20'
  if (leftHrs < 2) css = 'text-reach bg-reach/10 border-reach/20'
  else if (leftHrs < 4) css = 'text-match bg-match/10 border-match/20'

  return { text, class: css }
}
</script>

<template>
  <div class="px-7 py-6 max-w-[1200px] mx-auto space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-xl font-bold text-text">Review Queue</h2>
        <p class="text-[13px] text-text-muted mt-1">Cases flagged by AI needing human escalation.</p>
      </div>
      <div>
        <button @click="fetchQueue" class="p-2 border border-black/10 rounded-lg hover:bg-bg transition-colors" title="Refresh">
          <svg class="w-4 h-4 text-text-muted" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path></svg>
        </button>
      </div>
    </div>

    <!-- Active List -->
    <div class="bg-surface rounded-xl border border-black/5 shadow-sm overflow-hidden">
      <div v-if="loading" class="p-10 text-center text-text-muted text-[13px]">Loading queue...</div>
      <div v-else-if="queue.length === 0" class="p-12 flex flex-col items-center justify-center text-center">
        <div class="w-16 h-16 bg-safe/10 rounded-full flex items-center justify-center mb-4">
          <svg class="w-8 h-8 text-safe" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
        </div>
        <h3 class="text-[15px] font-bold text-text mb-1">Queue is empty!</h3>
        <p class="text-[13px] text-text-muted">All escalated cases have been cleared.</p>
      </div>

      <div v-else class="divide-y divide-black/5">
        <div 
          v-for="c in queue" 
          :key="c.id"
          class="p-5 flex items-center justify-between hover:bg-bg/50 transition-colors group"
        >
          <div class="flex items-center gap-5 w-1/3">
            <div class="w-10 h-10 rounded-full bg-secondary text-primary font-medium flex items-center justify-center shrink-0 border border-primary/10">
              {{ getAvatar(c.student?.full_name) }}
            </div>
            <div>
              <div class="font-bold text-text mb-0.5">{{ c.student?.full_name }}</div>
              <div class="text-[11px] text-text-muted">Case ID: <span class="font-mono text-[10px]">{{ c.id.substring(0,8) }}</span></div>
            </div>
          </div>

          <div class="w-1/3 px-4 border-l border-black/5">
            <div class="text-[12px] font-medium text-reach mb-1 flex items-center gap-1.5">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
              Escalation Reason
            </div>
            <div class="text-[13px] text-text-muted line-clamp-2">
              {{ c.escalation_reason || 'AI failed to determine definite matching strategy.' }}
            </div>
          </div>

          <div class="w-1/3 flex justify-end gap-6 items-center">
             <div class="text-right">
               <div class="text-[11px] text-text-muted uppercase tracking-wider mb-1 mt-0.5">SLA Deadline</div>
               <span class="px-2 py-0.5 rounded text-[11px] font-bold border" :class="getSlaInfo(c.created_at).class">
                 {{ getSlaInfo(c.created_at).text }}
               </span>
             </div>
             <button @click="router.push('/cases/' + c.id)" class="px-5 py-2 text-[13px] font-medium bg-primary text-white hover:bg-primary-hover rounded-lg transition-colors shadow-sm whitespace-nowrap">
               Review Case
             </button>
          </div>

        </div>
      </div>
    </div>
  </div>
</template>
