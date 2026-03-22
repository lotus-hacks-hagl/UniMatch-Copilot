<script setup>
import { onMounted, computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useCasesStore } from '../stores/casesStore'
import { useAuthStore } from '../stores/authStore'
import { useI18n } from 'vue-i18n'
import { useToast } from '../composables/useToast'
import { formatFloat2 } from '../utils/number'

import { Chart as ChartJS, Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale, ArcElement, LineElement, PointElement, Filler } from 'chart.js'
import { Bar, Doughnut, Line } from 'vue-chartjs'

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend, ArcElement, LineElement, PointElement, Filler)

const router = useRouter()
const casesStore = useCasesStore()
const authStore = useAuthStore()
const { cases, stats, filter: activeFilter } = storeToRefs(casesStore)
const { t } = useI18n()
const toast = useToast()
const searchTerm = ref('')

const filteredCases = computed(() => {
  const term = searchTerm.value.trim().toLowerCase()
  if (!term) return cases.value
  return (cases.value || []).filter((c) => {
    const s = c.student || {}
    const hay = [
      s.full_name,
      s.intended_major,
      s.target_intake,
      s.preferred_countries ? s.preferred_countries.join(' ') : ''
    ]
      .filter(Boolean)
      .join(' ')
      .toLowerCase()
    return hay.includes(term)
  })
})

const filters = ['All cases', 'Done', 'Processing', 'Human review']

const handleClaim = async (event, id) => {
  event.stopPropagation()
  try {
    await casesStore.claimCase(id)
    toast.addToast(t('dialogs.caseClaimed'), 'success')
  } catch (err) {
    toast.addToast(t('dialogs.claimFailed'), 'error')
  }
}

// Chart reactive data
const barChartData = ref({ labels: [], datasets: [] })
const donutChartData = ref({ labels: [], datasets: [] })
const lineChartData = ref({ labels: [], datasets: [] })

const barChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { display: false } },
  scales: { 
    x: { grid: { display: false } }, 
    y: { grid: { borderDash: [4, 4] }, beginAtZero: true } 
  }
}

const donutChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  cutout: '75%',
  plugins: { 
    legend: { position: 'right', labels: { usePointStyle: true, padding: 20 } } 
  }
}

const lineChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { display: false } },
  scales: { 
    x: { grid: { display: false } }, 
    y: { grid: { borderDash: [4, 4] }, beginAtZero: true } 
  }
}

const changeFilter = (f) => {
  casesStore.fetchCases(f)
}

const showFilterComingSoon = () => {
  toast.addToast('Advanced filters coming soon', 'info')
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

const formatRelativeTime = (date) => {
  if (!date) return ''
  const now = new Date()
  const then = new Date(date)
  const diff = Math.floor((now - then) / 1000)
  if (diff < 60) return 'Just now'
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  return then.toLocaleDateString()
}

const loadCharts = async () => {
  const [dailyResponse, escalationResponse] = await Promise.all([
    casesStore.fetchCasesByDay(),
    casesStore.fetchEscalationTrend()
  ])

  // Process Bar Chart
  const dailyLabels = dailyResponse.map(d => new Date(d.date).toLocaleDateString(undefined, { weekday: 'short' }))
  const dailyCounts = dailyResponse.map(d => d.count)
  barChartData.value = {
    labels: dailyLabels.length ? dailyLabels : ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
    datasets: [{
      label: 'Cases',
      backgroundColor: '#a32d2d',
      borderRadius: 6,
      data: dailyCounts.length ? dailyCounts : [0, 0, 0, 0, 0, 0, 0]
    }]
  }

  // Process Line Chart
  const escalationLabels = escalationResponse.map(d => new Date(d.date).toLocaleDateString(undefined, { month: 'short', day: 'numeric' }))
  const escalationCounts = escalationResponse.map(d => d.count)
  lineChartData.value = {
    labels: escalationLabels.length ? escalationLabels : ['W1', 'W2', 'W3', 'W4', 'W5'],
    datasets: [{
      label: 'Escalations',
      borderColor: '#a32d2d',
      backgroundColor: 'rgba(163, 45, 45, 0.1)',
      data: escalationCounts.length ? escalationCounts : [0, 0, 0, 0, 0],
      fill: true,
      tension: 0.4
    }]
  }

  // Handle Donut (Static for now until specific distribution endpoint is added or client-side derived)
  const safeCount = cases.value.filter(c => c.status === 'done').length
  const reviewCount = cases.value.filter(c => c.status === 'human_review').length
  const procCount = cases.value.filter(c => c.status === 'processing').length
  
  donutChartData.value = {
    labels: ['Safe', 'Match', 'Reach'],
    datasets: [{
      backgroundColor: ['#66bb6a', '#f57f17', '#e0e0e0'],
      data: [safeCount, procCount, reviewCount], // Using status as a proxy for tier for now
      borderWidth: 0,
      hoverOffset: 6
    }]
  }
}

onMounted(async () => {
  await Promise.all([
    casesStore.fetchStats(),
    casesStore.fetchCases('All cases')
  ])
  await loadCharts()
})
</script>

<template>
  <div class="px-8 py-6 w-full max-w-[1600px] mx-auto space-y-8 font-sans">
    
    <!-- TOP KPIs ROW (Always Visible) -->
    <div class="grid grid-cols-4 gap-5">
      <div data-testid="cases-stats-cases-today" class="bg-white rounded-[20px] p-6 shadow-[0_4px_20px_rgba(0,0,0,0.03)] border border-black/[0.03] hover:-translate-y-1 hover:shadow-[0_8px_30px_rgba(0,0,0,0.06)] transition-all duration-300">
        <div class="text-[14px] font-medium text-[#18180f] mb-3">{{ $t('cases.casesToday') }}</div>
        <div class="text-[36px] font-bold text-[#18180f] leading-none text-transparent bg-clip-text bg-gradient-to-r from-[#ce3e3e] to-[#a32d2d]">{{ stats.casesToday || 0 }}</div>
      </div>
      <div data-testid="cases-stats-avg-processing" class="bg-white rounded-[20px] p-6 shadow-[0_4px_20px_rgba(0,0,0,0.03)] border border-black/[0.03] hover:-translate-y-1 hover:shadow-[0_8px_30px_rgba(0,0,0,0.06)] transition-all duration-300">
        <div class="text-[14px] font-medium text-[#18180f] mb-3">{{ $t('cases.avgProcessing') }}</div>
        <div class="text-[36px] font-bold text-[#18180f] leading-none">{{ stats.avgProcessingTime || '0m' }}</div>
      </div>
      <div data-testid="cases-stats-awaiting-review" class="bg-white rounded-[20px] p-6 shadow-[0_4px_20px_rgba(0,0,0,0.03)] border border-black/[0.03] hover:-translate-y-1 hover:shadow-[0_8px_30px_rgba(0,0,0,0.06)] transition-all duration-300 relative overflow-hidden">
        <div class="text-[14px] font-medium text-[#18180f] mb-3 flex items-center justify-between relative z-10">
          {{ $t('cases.awaitingReview') }}
          <span v-if="stats.awaitingReview > 0" class="w-2.5 h-2.5 rounded-full bg-red-500 animate-pulse"></span>
        </div>
        <div class="text-[36px] font-bold text-[#18180f] leading-none relative z-10">{{ stats.awaitingReview || 0 }}</div>
        <!-- subtle danger bg if high -->
        <div v-if="stats.awaitingReview > 0" class="absolute right-0 bottom-0 w-32 h-32 bg-red-500/5 rounded-full blur-2xl -mr-10 -mb-10 pointer-events-none"></div>
      </div>
      <div data-testid="cases-stats-ai-confidence" class="bg-white rounded-[20px] p-6 shadow-[0_4px_20px_rgba(0,0,0,0.03)] border border-black/[0.03] hover:-translate-y-1 hover:shadow-[0_8px_30px_rgba(0,0,0,0.06)] transition-all duration-300">
        <div class="text-[14px] font-medium text-[#18180f] mb-3">{{ $t('cases.aiConfidence') }}</div>
        <div class="text-[36px] font-bold text-[#18180f] leading-none flex items-end gap-2">
          {{ formatConfidence(stats.aiConfidenceAvg) }}%
          <svg v-if="formatConfidence(stats.aiConfidenceAvg) >= 80" class="w-6 h-6 text-[#2e7d32] mb-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"></path></svg>
        </div>
      </div>
    </div>

    <!-- TABS ROW -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-1 bg-white p-1 rounded-xl shadow-[0_4px_10px_rgba(0,0,0,0.02)] border border-black/5">
        <button 
          v-for="f in filters" 
          :key="f"
          @click="changeFilter(f)"
          class="px-5 py-2 text-[14px] font-bold rounded-lg transition-all duration-300"
          :class="activeFilter === f ? 'bg-[#a32d2d] shadow-sm text-white' : 'text-[#6b6a62] hover:text-[#18180f] hover:bg-black/5'"
        >
          {{ $t('cases.filters.' + f) }}
        </button>
      </div>
      <div class="flex items-center gap-3">
        <div class="relative shadow-sm rounded-lg">
          <input v-model="searchTerm" type="text" :placeholder="$t('cases.searchPlaceholder')" class="pl-10 pr-4 py-2 text-[14px] bg-white border border-black/10 focus:border-[#a32d2d] focus:ring-2 focus:ring-[#a32d2d]/10 outline-none rounded-xl w-[280px] transition-all" />
          <svg class="w-4 h-4 text-[#a8a79d] absolute left-3.5 top-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path></svg>
        </div>
        <button @click="showFilterComingSoon" class="p-2.5 bg-white rounded-xl text-[#6b6a62] border border-black/10 shadow-sm hover:bg-gray-50 hover:-translate-y-0.5 transition-all outline-none focus:ring-2 focus:ring-[#a32d2d]/10">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"></path></svg>
        </button>
      </div>
    </div>

    <!-- MAIN DASHBOARD CONTENT -->
    <div class="flex flex-col xl:flex-row gap-6">
      
      <!-- DATA TABLE AREA -->
      <div class="flex-1 bg-white rounded-[20px] shadow-[0_4px_20px_rgba(0,0,0,0.03)] border border-black/[0.03] overflow-hidden min-h-[460px] flex flex-col">
        <div v-if="casesStore.loading" class="flex items-center justify-center flex-1 text-[#6b6a62] p-12">
          <div class="w-8 h-8 border-4 border-red-100 border-t-[#a32d2d] rounded-full animate-spin"></div>
        </div>
        
        <table v-else-if="filteredCases.length > 0" class="w-full text-left border-collapse">
          <thead>
            <tr class="text-[12px] text-[#8a8980] uppercase tracking-wider border-b border-black/5 bg-[#fafafa]">
              <th class="px-6 py-4 font-bold">{{ $t('cases.table.student') }}</th>
              <th class="px-6 py-4 font-bold">{{ $t('cases.table.target') }}</th>
              <th class="px-6 py-4 font-bold">{{ $t('cases.table.aiMatch') }}</th>
              <th class="px-6 py-4 font-bold">{{ $t('cases.table.confidence') }}</th>
              <th class="px-6 py-4 font-bold">{{ $t('cases.table.status') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-black/5 text-[14px]">
            <tr 
              v-for="c in filteredCases" 
              :key="c.id"
              @click="router.push('/cases/' + c.id)"
              class="hover:bg-gray-50 transition-colors group cursor-pointer"
            >
              <td class="px-6 py-4">
                <div class="flex items-center gap-3">
                  <div class="w-9 h-9 rounded-full bg-[#f4f5f7] text-[#18180f] font-bold flex items-center justify-center shrink-0 border border-black/5">
                    {{ getAvatar(c.student?.full_name) }}
                  </div>
                  <div>
                    <div class="font-bold text-[#18180f] group-hover:text-[#a32d2d] transition-colors">{{ c.student?.full_name }}</div>
                    <div class="text-[12px] text-[#6b6a62] mt-0.5" v-if="c.student">GPA {{ formatFloat2(c.student.gpa_normalized, 'N/A') }} • IELTS {{ c.student.ielts_overall ? formatFloat2(c.student.ielts_overall, 'N/A') : 'N/A' }}</div>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4">
                <div class="text-[#18180f] font-bold">{{ c.student?.intended_major }}</div>
                <div class="text-[12px] text-[#6b6a62] mt-0.5">{{ c.student?.target_intake }}</div>
              </td>
              <td class="px-6 py-4">
                <div class="flex gap-2" v-if="c.recommendations">
                  <span v-if="getTiers(c.recommendations).safe > 0" class="px-2.5 py-1 rounded text-[11px] font-bold bg-[#e8f5e9] text-[#2e7d32]">{{ getTiers(c.recommendations).safe }} {{ $t('cases.tiers.safe') }}</span>
                  <span v-else-if="getTiers(c.recommendations).reach > 0" class="px-2.5 py-1 rounded text-[11px] font-bold bg-red-50 text-[#a32d2d]">{{ getTiers(c.recommendations).reach }} {{ $t('cases.tiers.reach') }}</span>
                </div>
                <div v-else class="text-[#8a8980] text-[12px] italic">{{ $t('cases.pendingAi') }}</div>
              </td>
              <td class="px-6 py-4">
                <span class="text-[13px] font-bold" :class="formatConfidence(c.ai_confidence) >= 90 ? 'text-[#2e7d32]' : (formatConfidence(c.ai_confidence) >= 80 ? 'text-[#f57f17]' : 'text-[#a32d2d]')">{{ formatConfidence(c.ai_confidence) }}%</span>
              </td>
              <td class="px-6 py-4">
                <span class="inline-flex items-center px-2.5 py-1 rounded-md text-[11px] font-bold" :class="getStatusClass(c.status)">
                  {{ getStatusLabel(c.status) }}
                </span>
                <div class="text-[11px] text-[#8a8980] mt-1">{{ formatRelativeTime(c.created_at) }}</div>
              </td>
            </tr>
          </tbody>
        </table>
        
        <!-- Smart Empty State contained IN the table card -->
        <div v-else class="flex flex-col items-center justify-center flex-1 p-16 text-center bg-[#fafafa] border border-black/5 mx-8 my-8 rounded-3xl shrink-0">
          <div class="relative mb-5 text-[#a32d2d] bg-red-50 p-6 rounded-full">
            <svg class="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path></svg>
          </div>
          <h3 class="text-[18px] font-bold text-[#18180f] mb-1">{{ $t('cases.noCases') }}</h3>
          <p class="text-[14px] text-[#6b6a62] max-w-sm mx-auto">{{ $t('cases.analytics.noCasesSub') || 'Try adjusting your filters or search terms to find what you are looking for.' }}</p>
          <button @click="router.push('/cases/new')" class="mt-6 px-6 py-3 bg-[#a32d2d] text-white rounded-xl text-[14px] font-bold shadow-[0_4px_14px_rgba(163,45,45,0.35)] hover:shadow-[0_6px_20px_rgba(163,45,45,0.5)] hover:-translate-y-0.5 transition-all">
            {{ $t('common.newCase') }}
          </button>
        </div>
      </div>

      <!-- CHARTS SIDEBAR -->
      <div class="w-full xl:w-[420px] shrink-0 space-y-6 flex flex-col justify-start">
        
        <!-- Cases per day Chart -->
        <div data-testid="cases-chart-cases-per-day" class="bg-white rounded-[20px] p-6 shadow-[0_4px_20px_rgba(0,0,0,0.03)] border border-black/[0.03]">
          <div class="text-[15px] font-bold text-[#18180f] mb-4">{{ $t('cases.charts.casesPerDay') }}</div>
          <div class="h-[180px] w-full">
            <Bar :data="barChartData" :options="barChartOptions" />
          </div>
        </div>
        
        <!-- Match Tier Distribution Chart -->
        <div data-testid="cases-chart-match-tier" class="bg-white rounded-[20px] p-6 shadow-[0_4px_20px_rgba(0,0,0,0.03)] border border-black/[0.03]">
          <div class="text-[15px] font-bold text-[#18180f] mb-2">{{ $t('cases.charts.matchTierDist') }}</div>
          <!-- Absolute centered total counter -->
          <div class="relative h-[180px] w-full flex items-center justify-center">
            <div class="absolute inset-0 flex flex-col items-center justify-center pointer-events-none pb-2 pr-16 text-center">
              <span class="text-[28px] font-black text-[#18180f] leading-none">{{ cases.length }}</span>
              <span class="text-[11px] font-bold text-[#8a8980] uppercase tracking-wider">Total</span>
            </div>
            <Doughnut :data="donutChartData" :options="donutChartOptions" />
          </div>
        </div>
        
        <!-- Escalation Rate Trend Chart -->
        <div data-testid="cases-chart-escalation-trend" class="bg-white rounded-[20px] p-6 shadow-[0_4px_20px_rgba(0,0,0,0.03)] border border-black/[0.03]">
          <div class="text-[15px] font-bold text-[#18180f] mb-4">{{ $t('cases.charts.escalationRate') }}</div>
          <div class="h-[140px] w-full">
            <Line :data="lineChartData" :options="lineChartOptions" />
          </div>
        </div>

      </div>

    </div>
  </div>
</template>
