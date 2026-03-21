<script setup>
import { onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useCasesStore } from '../stores/casesStore'
import { useAuthStore } from '../stores/authStore'
import { useI18n } from 'vue-i18n'

const router = useRouter()
const casesStore = useCasesStore()
const authStore = useAuthStore()
const { cases, stats, filter: activeFilter } = storeToRefs(casesStore)
const { t } = useI18n()

const filters = ['All cases', 'Done', 'Processing', 'Human review']

const handleClaim = async (event, id) => {
  event.stopPropagation()
  try {
    await casesStore.claimCase(id)
    alert('Case claimed successfully!')
  } catch (err) {
    alert('Failed to claim case')
  }
}

onMounted(() => {
  casesStore.fetchStats()
  casesStore.fetchCases('All cases')
})

const changeFilter = (f) => {
  casesStore.fetchCases(f)
}

const getStatusClass = (statusStr) => {
  const status = (statusStr || '').toLowerCase()
  if (status === 'done') return 'bg-safe/10 text-safe border-safe/20'
  if (status === 'processing') return 'bg-match/10 text-match border-match/20'
  return 'bg-reach/10 text-reach border-reach/20'
}

const getStatusLabel = (statusStr) => {
  const status = (statusStr || '').toLowerCase()
  if (status === 'human_review') return t('cases.filters.Human review')
  if (status === 'done') return t('cases.filters.Done')
  if (status === 'processing') return t('cases.filters.Processing')
  return statusStr || t('cases.pendingAi')
}

const getAvatar = (name) => {
  if (!name) return '??'
  const parts = name.split(' ')
  if (parts.length >= 2) return (parts[0][0] + parts[parts.length-1][0]).toUpperCase()
  return name.substring(0, 2).toUpperCase()
}

const formatBudget = (val) => {
  if (!val) return t('cases.noBudget')
  return `$${Math.round(val/1000)}k/yr`
}

const getTiers = (recommendations) => {
  if (!recommendations) return {}
  const res = { safe: 0, match: 0, reach: 0 }
  recommendations.forEach(r => {
    if (r.tier === 'safe') res.safe++
    if (r.tier === 'match') res.match++
    if (r.tier === 'reach') res.reach++
  })
  return res
}

const formatConfidence = (val) => {
  return Math.round((val || 0) * 100)
}

const formatRelativeTime = (dateStr) => {
  if (!dateStr) return t('cases.unknown')
  const diff = Date.now() - new Date(dateStr).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 60) return `${mins}${t('cases.m')} ${t('cases.ago')}`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}${t('cases.h')} ${t('cases.ago')}`
  return `${Math.floor(hrs / 24)}${t('cases.d')} ${t('cases.ago')}`
}
</script>

<template>
  <div class="px-7 py-6 max-w-[1400px] mx-auto space-y-6">
    
    <!-- Metrics Row -->
    <div class="grid grid-cols-4 gap-4">
      <div class="bg-surface rounded-xl p-5 border border-black/5 shadow-sm relative overflow-hidden group">
        <div class="absolute inset-y-0 left-0 w-1 bg-primary/0 group-hover:bg-primary transition-colors"></div>
        <div class="text-[13px] text-text-muted mb-1 flex items-center justify-between">{{ $t('cases.casesToday') }}</div>
        <div class="text-2xl font-bold text-text mb-1">{{ stats.casesToday || 0 }}</div>
      </div>
      <div class="bg-surface rounded-xl p-5 border border-black/5 shadow-sm relative overflow-hidden group">
        <div class="absolute inset-y-0 left-0 w-1 bg-primary/0 group-hover:bg-primary transition-colors"></div>
        <div class="text-[13px] text-text-muted mb-1 flex items-center justify-between">{{ $t('cases.avgProcessing') }}</div>
        <div class="text-2xl font-bold text-text mb-1">{{ stats.avgProcessingTime || '0m' }}</div>
      </div>
      <div class="bg-surface rounded-xl p-5 border border-black/5 shadow-sm relative overflow-hidden group">
        <div class="absolute inset-y-0 left-0 w-1 bg-primary/0 group-hover:bg-primary transition-colors"></div>
        <div class="text-[13px] text-text-muted mb-1 flex items-center justify-between">
          {{ $t('cases.awaitingReview') }}
          <span v-if="stats.awaitingReview > 0" class="w-2 h-2 rounded-full bg-red-500 animate-pulse"></span>
        </div>
        <div class="text-2xl font-bold text-text mb-1">{{ stats.awaitingReview || 0 }}</div>
      </div>
      <div class="bg-surface rounded-xl p-5 border border-black/5 shadow-sm relative overflow-hidden group">
        <div class="absolute inset-y-0 left-0 w-1 bg-primary/0 group-hover:bg-primary transition-colors"></div>
        <div class="text-[13px] text-text-muted mb-1 flex items-center justify-between">{{ $t('cases.aiConfidence') }}</div>
        <div class="text-2xl font-bold text-text mb-1">{{ formatConfidence(stats.aiConfidenceAvg) }}%</div>
      </div>
    </div>

    <!-- Main Table Section -->
    <div class="bg-surface rounded-xl border border-black/5 shadow-sm overflow-hidden flex flex-col">
      <!-- Tabs & Actions -->
      <div class="border-b border-black/5 px-5 flex items-center justify-between">
        <div class="flex items-center gap-6">
          <button 
            v-for="f in filters" 
            :key="f"
            @click="changeFilter(f)"
            class="py-3.5 text-[13px] font-medium border-b-2 transition-colors"
            :class="activeFilter === f ? 'border-primary text-primary' : 'border-transparent text-text-muted hover:text-text'"
          >
            {{ $t('cases.filters.' + f) }}
          </button>
        </div>
        <div class="flex items-center gap-3">
          <div class="relative">
            <input type="text" :placeholder="$t('cases.searchPlaceholder')" class="pl-8 pr-3 py-1.5 text-[13px] bg-bg border-transparent focus:border-primary focus:ring-1 focus:ring-primary rounded-lg w-[200px]" />
            <svg class="w-4 h-4 text-text-muted absolute left-2.5 top-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path></svg>
          </div>
          <button class="p-1.5 rounded text-text-muted border border-black/10 hover:bg-bg transition-colors">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"></path></svg>
          </button>
        </div>
      </div>

      <!-- Table -->
      <div class="overflow-x-auto min-h-[300px]">
        <div v-if="casesStore.loading" class="flex items-center justify-center h-[300px] text-text-muted">{{ $t('cases.loading') }}</div>
        <table v-else-if="cases.length > 0" class="w-full text-left border-collapse">
          <thead>
            <tr class="text-[11px] text-text-muted uppercase tracking-wider border-b border-black/5">
              <th class="px-5 py-3 font-medium">{{ $t('cases.table.student') }}</th>
              <th class="px-5 py-3 font-medium">{{ $t('cases.table.profile') }}</th>
              <th class="px-5 py-3 font-medium">{{ $t('cases.table.target') }}</th>
              <th class="px-5 py-3 font-medium">{{ $t('cases.table.aiMatch') }}</th>
              <th class="px-5 py-3 font-medium cursor-pointer flex items-center gap-1 hover:text-text hover:underline decoration-primary/20 underline-offset-4">{{ $t('cases.table.confidence') }} <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg></th>
              <th class="px-5 py-3 font-medium">{{ $t('cases.table.status') }}</th>
              <th class="px-5 py-3 font-medium">Assignee</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-black/5 text-[13px]">
            <tr 
              v-for="c in cases" 
              :key="c.id"
              @click="router.push('/cases/' + c.id)"
              class="hover:bg-bg/50 transition-colors group cursor-pointer"
            >
              <td class="px-5 py-3">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-full bg-secondary text-primary font-medium flex items-center justify-center shrink-0 border border-primary/10">
                    {{ getAvatar(c.student?.full_name) }}
                  </div>
                  <div>
                    <div class="font-medium text-text group-hover:text-primary transition-colors">{{ c.student?.full_name }}</div>
                    <div class="text-[11px] text-text-muted mt-0.5" v-if="c.student">GPA {{ c.student.gpa_normalized }} • IELTS {{ c.student.ielts_overall }}</div>
                  </div>
                </div>
              </td>
              <td class="px-5 py-3">
                <div class="text-text">{{ c.student?.intended_major }}</div>
                <div class="text-[11px] text-text-muted mt-0.5">{{ (c.student?.preferred_countries || []).join(', ') }} • {{ formatBudget(c.student?.budget_usd_per_year) }}</div>
              </td>
              <td class="px-5 py-3">
                <div class="text-text">{{ c.student?.target_intake }}</div>
                <div class="text-[11px] text-text-muted mt-0.5">Bachelor</div>
              </td>
              <td class="px-5 py-3">
                <div class="flex gap-1.5" v-if="c.recommendations">
                  <span v-if="getTiers(c.recommendations).safe > 0" class="px-2 py-0.5 rounded text-[10px] font-medium bg-safe/10 text-safe border border-safe/20">{{ getTiers(c.recommendations).safe }} {{ $t('cases.tiers.safe') }}</span>
                  <span v-if="getTiers(c.recommendations).match > 0" class="px-2 py-0.5 rounded text-[10px] font-medium bg-match/10 text-match border border-match/20">{{ getTiers(c.recommendations).match }} {{ $t('cases.tiers.match') }}</span>
                  <span v-if="getTiers(c.recommendations).reach > 0" class="px-2 py-0.5 rounded text-[10px] font-medium bg-reach/10 text-reach border border-reach/20">{{ getTiers(c.recommendations).reach }} {{ $t('cases.tiers.reach') }}</span>
                </div>
                <div v-else class="text-text-muted text-[11px]">{{ $t('cases.pendingAi') }}</div>
              </td>
              <td class="px-5 py-3">
                <div class="flex items-center gap-2">
                  <div class="w-16 h-1.5 bg-gray-100 rounded-full overflow-hidden">
                    <div 
                      class="h-full rounded-full" 
                      :class="formatConfidence(c.ai_confidence) >= 90 ? 'bg-safe' : (formatConfidence(c.ai_confidence) >= 80 ? 'bg-match' : 'bg-reach')"
                      :style="{ width: Math.max(5, formatConfidence(c.ai_confidence)) + '%' }"
                    ></div>
                  </div>
                  <span class="text-[11px] font-medium" :class="formatConfidence(c.ai_confidence) >= 90 ? 'text-safe' : (formatConfidence(c.ai_confidence) >= 80 ? 'text-match' : 'text-reach')">{{ formatConfidence(c.ai_confidence) }}%</span>
                </div>
              </td>
              <td class="px-5 py-3">
                <div class="flex items-center justify-between">
                  <span class="inline-flex items-center px-2 py-0.5 rounded-md text-[11px] font-medium border" :class="getStatusClass(c.status)">
                    {{ getStatusLabel(c.status) }}
                  </span>
                </div>
              </td>
              <td class="px-6 py-4">
                <div v-if="c.assigned_to" class="flex items-center text-xs text-[#6b6a62]">
                  <div class="w-5 h-5 bg-gray-200 rounded-full flex items-center justify-center mr-2 text-[10px] font-bold">
                    {{ c.assigned_to.username.charAt(0).toUpperCase() }}
                  </div>
                  {{ c.assigned_to.username }}
                </div>
                <button 
                  v-else-if="!authStore.isAdmin"
                  @click="handleClaim($event, c.id)"
                  class="bg-[#a32d2d] text-white px-3 py-1 rounded text-[11px] font-bold hover:bg-[#821419] transition-colors"
                >
                  Claim
                </button>
                <span v-else class="text-[11px] text-[#6b6a62] italic">Pool</span>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-else class="flex flex-col items-center justify-center p-12 h-[300px] bg-surface rounded-lg border-2 border-dashed border-black/10 mx-5 my-5">
          <svg class="w-10 h-10 text-black/20 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 002-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path></svg>
          <p class="text-[13px] text-text-muted">{{ $t('cases.noCases') }}</p>
        </div>
      </div>
    </div>

    <!-- Analytics Row -->
    <div class="grid grid-cols-3 gap-4">
      <!-- Empty charts placeholder to preserve UI layout -->
      <div class="bg-surface rounded-xl p-5 border border-black/5 shadow-sm">
        <div class="text-[13px] text-text-muted mb-4 font-medium">{{ $t('cases.charts.casesPerDay') }}</div>
        <div class="h-24 flex items-end gap-2 justify-between">
          <div class="w-full bg-primary/20 hover:bg-primary transition-colors rounded-t" style="height: 40%"></div>
          <div class="w-full bg-primary/20 hover:bg-primary transition-colors rounded-t" style="height: 60%"></div>
          <div class="w-full bg-primary/40 hover:bg-primary transition-colors rounded-t" style="height: 30%"></div>
          <div class="w-full bg-primary/80 hover:bg-primary transition-colors rounded-t" style="height: 90%"></div>
          <div class="w-full bg-primary hover:bg-primary transition-colors rounded-t" style="height: 100%"></div>
          <div class="w-full bg-primary/60 hover:bg-primary transition-colors rounded-t" style="height: 70%"></div>
          <div class="w-full bg-primary/40 hover:bg-primary transition-colors rounded-t" style="height: 50%"></div>
        </div>
      </div>
      <div class="bg-surface rounded-xl p-5 border border-black/5 shadow-sm">
        <div class="text-[13px] text-text-muted mb-4 font-medium flex justify-between">{{ $t('cases.charts.matchTierDist') }} <span class="text-safe">{{ $t('cases.tiers.safe') }}</span></div>
        <div class="h-24 flex items-center justify-center relative">
          <!-- Fake Donut -->
          <div class="w-20 h-20 rounded-full border-4 border-t-safe border-r-match border-b-reach border-l-gray-100 flex items-center justify-center">
             <span class="text-lg font-bold text-text">...</span>
          </div>
        </div>
      </div>
      <div class="bg-surface rounded-xl p-5 border border-black/5 shadow-sm">
        <div class="text-[13px] text-text-muted mb-4 font-medium">{{ $t('cases.charts.escalationRate') }}</div>
        <div class="h-24 relative overflow-hidden">
           <svg viewBox="0 0 100 40" class="w-full h-full preserve-aspect-ratio cursor-pointer group">
             <path d="M0,35 Q10,30 20,32 T40,20 T60,25 T80,10 T100,5" fill="none" class="stroke-reach opacity-50 group-hover:opacity-100 transition-opacity" stroke-width="2" stroke-linecap="round"/>
             <circle cx="100" cy="5" r="2" class="fill-reach" />
             <path d="M0,40 L0,35 Q10,30 20,32 T40,20 T60,25 T80,10 T100,5 L100,40 Z" class="fill-reach/10" />
           </svg>
        </div>
      </div>
    </div>
  </div>
</template>
