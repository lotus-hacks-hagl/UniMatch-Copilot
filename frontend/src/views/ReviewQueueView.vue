<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../services/api'
import { useI18n } from 'vue-i18n'

const router = useRouter()
const { t } = useI18n()
const queue = ref([])
const loading = ref(true)

const fetchQueue = async () => {
  loading.value = true
  try {
    const response = await api.get('/cases', {
      params: { status: 'human_review', page: 1, limit: 100 }
    })
    queue.value = response.data.data || []
  } catch (error) {
    console.error('Failed to fetch human review queue', error)
  } finally {
    loading.value = false
  }
}

onMounted(fetchQueue)

const getAvatar = (name) => {
  if (!name) return '??'
  const parts = name.split(' ')
  if (parts.length >= 2) return (parts[0][0] + parts[parts.length-1][0]).toUpperCase()
  return name.substring(0, 2).toUpperCase()
}

const getSlaInfo = (createdAt) => {
  const createdDate = new Date(createdAt)
  const deadlineMs = createdDate.getTime() + 8 * 3600 * 1000
  const leftMs = deadlineMs - Date.now()
  
  if (leftMs <= 0) return { text: t('reviewQueue.overdue'), class: 'text-[#a32d2d] bg-red-50 border-red-100' }
  const leftHrs = Math.floor(leftMs / 3600000)
  const leftMins = Math.floor((leftMs % 3600000) / 60000)
  
  const text = t('reviewQueue.timeLeft', { hrs: leftHrs, mins: leftMins })
  let css = 'text-[#2e7d32] bg-[#e8f5e9] border-[#c8e6c9]'
  if (leftHrs < 2) css = 'text-[#a32d2d] bg-red-50 border-red-100'
  else if (leftHrs < 4) css = 'text-[#f57f17] bg-[#fff8e1] border-[#ffecb3]'

  return { text, class: css }
}
</script>

<template>
  <div class="px-8 py-6 max-w-7xl mx-auto space-y-8 font-sans">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-2xl font-bold text-[#18180f] tracking-tight">{{ $t('reviewQueue.title') }}</h2>
        <p class="text-[14px] text-[#6b6a62] mt-1">{{ $t('reviewQueue.subtitle') }}</p>
      </div>
      <div>
        <button @click="fetchQueue" class="btn-outline">
          <svg class="w-4 h-4 text-[#6b6a62]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path></svg>
          Refresh Queue
        </button>
      </div>
    </div>

    <!-- Active List -->
    <div class="card-soft overflow-hidden p-0 flex flex-col min-h-[400px]">
      <Transition name="fade" mode="out-in">
        <div v-if="loading" class="flex-1 flex flex-col items-center justify-center p-12 space-y-4">
          <div class="w-10 h-10 border-4 border-red-100 border-t-[#a32d2d] rounded-full animate-spin"></div>
          <div class="text-[14px] font-medium text-[#6b6a62]">{{ $t('reviewQueue.loading') }}</div>
        </div>
        
        <div v-else-if="queue.length === 0" class="flex-1 flex flex-col items-center justify-center text-center p-12">
          <div class="w-20 h-20 bg-[#e8f5e9] rounded-full flex items-center justify-center mb-4">
            <svg class="w-10 h-10 text-[#2e7d32]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7"></path></svg>
          </div>
          <h3 class="text-[16px] font-bold text-[#18180f] mb-1">{{ $t('reviewQueue.empty') }}</h3>
          <p class="text-[14px] text-[#6b6a62]">{{ $t('reviewQueue.cleared') }}</p>
        </div>

        <div v-else class="w-full flex-1">
          <TransitionGroup name="list" tag="div" class="divide-y divide-black/5">
            <div 
              v-for="c in queue" 
              :key="c.id"
              class="p-6 flex items-center justify-between hover:bg-gray-50/80 transition-colors group cursor-pointer"
              @click="router.push('/cases/' + c.id)"
            >
            <!-- Student / Case -->
            <div class="flex items-center gap-4 w-1/3">
              <div class="w-10 h-10 rounded-full bg-[#f4f5f7] text-[#18180f] font-bold flex items-center justify-center shrink-0 border border-black/5">
                {{ getAvatar(c.student?.full_name) }}
              </div>
              <div>
                <div class="font-bold text-[#18180f] group-hover:text-[#a32d2d] transition-colors text-[15px] mb-0.5">{{ c.student?.full_name }}</div>
                <div class="text-[12px] text-[#6b6a62]">{{ $t('reviewQueue.caseId') }} <span class="font-mono text-[11px] font-bold">{{ c.id.substring(0,8) }}</span></div>
              </div>
            </div>

            <!-- Escalation Info -->
            <div class="w-1/3 px-6 border-l border-black/5">
              <div class="text-[12px] font-bold text-[#a32d2d] mb-1 flex items-center gap-1.5 uppercase tracking-wide">
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
                {{ $t('reviewQueue.escalationReason') }}
              </div>
              <div class="text-[13px] text-[#6b6a62] line-clamp-2">
                {{ c.escalation_reason || $t('reviewQueue.defaultReason') }}
              </div>
            </div>

            <!-- SLA & Action -->
            <div class="w-1/3 flex justify-end gap-6 items-center">
               <div class="text-right">
                 <div class="text-[11px] text-[#8a8980] uppercase tracking-wider font-bold mb-1">{{ $t('reviewQueue.slaDeadline') }}</div>
                 <span class="px-2.5 py-1 rounded-md text-[11px] font-bold border" :class="getSlaInfo(c.created_at).class">
                   {{ getSlaInfo(c.created_at).text }}
                 </span>
               </div>
               <button @click.stop="router.push('/cases/' + c.id)" class="btn-primary">
                 {{ $t('reviewQueue.reviewCase') }}
               </button>
            </div>
          </div>
          </TransitionGroup>
        </div>
      </Transition>
    </div>
  </div>
</template>
