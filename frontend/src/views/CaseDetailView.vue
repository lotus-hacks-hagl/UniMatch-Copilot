<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../services/api'
import { usePolling } from '../composables/usePolling'
import { useI18n } from 'vue-i18n'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const activeTab = ref('profile')
const tabs = ['profile', 'aiAnalysis', 'documents', 'communication']

const caseData = ref(null)
const loading = ref(true)

const fetchCase = async () => {
  try {
    const res = await api.get('/cases/' + route.params.id)
    caseData.value = res.data
  } catch (err) {
    console.error('Fetch case failed', err)
  } finally {
    loading.value = false
  }
}

// Polling setup if case is pending/processing
let pollTimer = null
onMounted(() => {
  fetchCase().then(() => {
    if (caseData.value && ['pending', 'processing'].includes(caseData.value.status)) {
      pollTimer = setInterval(async () => {
        await fetchCase()
        if (!['pending', 'processing'].includes(caseData.value?.status)) {
          clearInterval(pollTimer)
        }
      }, 3000)
    }
  })
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})

const getAvatar = (name) => {
  if (!name) return '?'
  return name.substring(0, 2).toUpperCase()
}

const formatConfidence = (val) => Math.round((val || 0) * 100)

const formatBudget = (val) => {
  if (!val) return t('caseDetail.noBudget')
  return `$${Math.round(val/1000)}k`
}

const generateReport = async () => {
  try {
    await api.post(`/cases/${route.params.id}/report`)
    alert('Report generation triggered successfully.')
  } catch (err) {
    console.error(err)
    alert('Failed to trigger report.')
  }
}

</script>

<template>
  <div class="px-7 py-6 max-w-[1400px] mx-auto">
    <div v-if="loading" class="text-text-muted mt-10 text-center">{{ $t('caseDetail.loading') }}</div>
    
    <div v-else-if="caseData" class="space-y-6">
      
      <!-- Meta & Breadcrumb -->
      <div class="flex items-center gap-3 text-[13px]">
        <router-link to="/cases" class="text-text-muted hover:text-text flex items-center gap-1 transition-colors">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"></path></svg>
          {{ $t('common.cases') }}
        </router-link>
        <span class="text-black/20">/</span>
        <span class="font-medium text-text">{{ caseData.student?.full_name || $t('caseDetail.student') }}</span>
        
        <span class="ml-4 px-2 py-0.5 rounded-md text-[11px] font-medium border uppercase tracking-wider"
              :class="{
                'bg-safe/10 text-safe border-safe/20': caseData.status === 'done',
                'bg-match/10 text-match border-match/20': caseData.status === 'processing',
                'bg-reach/10 text-reach border-reach/20': caseData.status === 'human_review',
                'bg-gray-100 text-text-muted border-black/10': !['done','processing','human_review'].includes(caseData.status)
              }">
          {{ (caseData.status || 'pending').replace('_', ' ') }}
        </span>
      </div>

      <!-- Layout: Profile Left (300px), Details Right -->
      <div class="flex xl:flex-row flex-col gap-6 items-start">
        
        <!-- Left: Profile Snapshot -->
        <div class="w-full xl:w-[320px] shrink-0 space-y-4">
          
          <!-- Confidence Card -->
          <div class="bg-surface rounded-xl p-5 border border-primary/20 shadow-sm relative overflow-hidden">
            <div class="absolute inset-x-0 top-0 h-1" 
                 :class="formatConfidence(caseData.ai_confidence) >= 90 ? 'bg-safe' : (formatConfidence(caseData.ai_confidence)>=80 ? 'bg-match' : 'bg-reach')"></div>
            <div class="flex items-start justify-between mb-4">
              <div class="text-[13px] font-medium text-text mt-1">{{ $t('caseDetail.aiMatchVerdict') }}</div>
              <div class="text-3xl font-bold" :class="formatConfidence(caseData.ai_confidence)>=90?'text-safe':(formatConfidence(caseData.ai_confidence)>=80?'text-match':'text-reach')">
                {{ formatConfidence(caseData.ai_confidence) }}%
              </div>
            </div>
            <p class="text-[13px] text-text-muted leading-relaxed">
              {{ caseData.status === 'human_review' ? caseData.escalation_reason || $t('caseDetail.escalatedReview') : $t('caseDetail.autoApproved') }}
            </p>
          </div>

          <!-- Student Card -->
          <div class="bg-surface rounded-xl p-5 border border-black/5 shadow-sm">
            <div class="flex items-center gap-4 border-b border-black/5 pb-5 mb-5">
              <div class="w-14 h-14 rounded-full bg-secondary text-primary text-xl font-bold flex items-center justify-center shrink-0">
                {{ getAvatar(caseData.student?.full_name) }}
              </div>
              <div>
                <h2 class="text-lg font-bold text-text mb-0.5">{{ caseData.student?.full_name }}</h2>
                <p class="text-[13px] text-text-muted">{{ caseData.student?.intended_major }} • {{ caseData.student?.target_intake }}</p>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-y-5 gap-x-4">
              <div>
                <div class="text-[11px] text-text-muted mb-0.5 uppercase tracking-wider">{{ $t('caseDetail.gpaNorm') }}</div>
                <div class="text-[14px] font-bold text-text">{{ caseData.student?.gpa_normalized || 'N/A' }}</div>
              </div>
              <div>
                <div class="text-[11px] text-text-muted mb-0.5 uppercase tracking-wider">{{ $t('caseDetail.ielts') }}</div>
                <div class="text-[14px] font-bold text-text">{{ caseData.student?.ielts_overall || 'N/A' }}</div>
              </div>
              <div>
                <div class="text-[11px] text-text-muted mb-0.5 uppercase tracking-wider">{{ $t('caseDetail.budget') }}</div>
                <div class="text-[14px] font-bold text-text">{{ formatBudget(caseData.student?.budget_usd_per_year) }}</div>
              </div>
              <div>
                <div class="text-[11px] text-text-muted mb-0.5 uppercase tracking-wider">{{ $t('caseDetail.countries') }}</div>
                <div class="text-[14px] font-bold text-text truncate max-w-[120px]" :title="(caseData.student?.preferred_countries||[]).join(', ')">
                  {{ (caseData.student?.preferred_countries||[]).join(', ') || $t('caseDetail.any') }}
                </div>
              </div>
            </div>

            <!-- Profile Summary Tags -->
            <div v-if="caseData.profile_summary" class="mt-6 pt-5 border-t border-black/5 space-y-4">
               <div>
                 <div class="text-[11px] text-text-muted mb-2 uppercase tracking-wider">{{ $t('caseDetail.strengths') }}</div>
                 <div class="flex flex-wrap gap-1.5">
                   <span v-for="(str, i) in caseData.profile_summary.strengths || []" :key="i" class="px-2 py-1 rounded bg-safe/10 text-safe text-[11px] font-medium border border-safe/20">
                     {{ str }}
                   </span>
                 </div>
               </div>
               <div>
                 <div class="text-[11px] text-text-muted mb-2 uppercase tracking-wider">{{ $t('caseDetail.risks') }}</div>
                 <div class="flex flex-wrap gap-1.5">
                   <span v-for="(r, i) in caseData.profile_summary.weaknesses || []" :key="i" class="px-2 py-1 rounded bg-reach/10 text-reach text-[11px] font-medium border border-reach/20">
                     {{ r }}
                   </span>
                 </div>
               </div>
            </div>

          </div>

          <div class="pt-4 flex flex-col gap-2">
            <button @click="generateReport" class="w-full py-2 bg-primary text-white rounded-lg text-[13px] font-medium hover:bg-primary-hover shadow-sm transition-colors flex items-center justify-center gap-2">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path></svg>
              {{ $t('caseDetail.generateReport') }}
            </button>
          </div>

        </div>

        <!-- Right: Tabs & Main Details -->
        <div class="flex-1 w-full bg-surface rounded-xl border border-black/5 shadow-sm min-h-[600px] flex flex-col">
          
          <div class="flex border-b border-black/5 px-4 pt-2 overflow-x-auto">
            <button 
              v-for="t in tabs" 
              :key="t"
              @click="activeTab = t"
              class="px-5 py-3 text-[13px] font-medium border-b-2 whitespace-nowrap transition-colors"
              :class="activeTab === t ? 'border-primary text-primary' : 'border-transparent text-text-muted hover:text-text'"
            >
              {{ $t('caseDetail.tabs.' + t) }}
            </button>
          </div>

          <div class="p-6 flex-1 bg-bg/20">
            <!-- Profile Tab Placeholder -->
            <div v-if="activeTab === 'profile'">
              <div class="bg-surface rounded-xl p-6 border border-black/10">
                <h3 class="text-base font-bold text-text mb-4">{{ $t('caseDetail.fullBackground') }}</h3>
                <div class="space-y-6 text-[13px]">
                  <div>
                    <div class="font-medium text-text mb-1 relative pb-2"><span class="absolute bottom-0 left-0 w-8 h-0.5 bg-primary/30"></span>{{ $t('caseDetail.extracurriculars') }}</div>
                    <p class="text-text-muted leading-relaxed whitespace-pre-wrap">{{ caseData.student?.extracurriculars || $t('caseDetail.noneRecorded') }}</p>
                  </div>
                  <div>
                    <div class="font-medium text-text mb-1 relative pb-2"><span class="absolute bottom-0 left-0 w-8 h-0.5 bg-primary/30"></span>{{ $t('caseDetail.achievements') }}</div>
                    <p class="text-text-muted leading-relaxed whitespace-pre-wrap">{{ caseData.student?.achievements || $t('caseDetail.noneRecorded') }}</p>
                  </div>
                </div>
              </div>
            </div>

            <!-- AI Analysis Tab (Recommendations list) -->
            <div v-if="activeTab === 'aiAnalysis'">
              <div v-if="['pending', 'processing'].includes(caseData.status)" class="flex flex-col items-center justify-center h-[300px]">
                <div class="w-8 h-8 rounded-full border-2 border-primary border-t-transparent animate-spin mb-4"></div>
                <div class="text-[13px] font-medium text-primary">{{ $t('caseDetail.aiAnalyzing') }}</div>
              </div>
              <div v-else-if="!caseData.recommendations || caseData.recommendations.length === 0" class="text-center mt-10 text-text-muted text-[13px]">
                {{ $t('caseDetail.noRecommendations') }}
              </div>
              <div v-else class="space-y-4">
                <div v-for="rec in [...caseData.recommendations].sort((a,b)=>a.rank_order - b.rank_order)" :key="rec.id" 
                     class="bg-surface p-5 rounded-xl border border-black/10 hover:border-primary/30 transition-colors shadow-sm relative"
                     :class="(rec.rank_order===1 && rec.tier==='match') ? 'ring-1 ring-primary/50' : ''">
                  
                  <div class="flex justify-between items-start mb-3">
                    <div>
                      <h4 class="font-bold text-text text-base">{{ rec.university_name }}</h4>
                      <p class="text-[12px] text-text-muted mt-0.5 line-clamp-2 pr-8">{{ rec.reason }}</p>
                    </div>
                    <span class="px-2.5 py-1 text-[11px] font-bold uppercase tracking-wider rounded border"
                          :class="rec.tier === 'safe' ? 'bg-safe/10 text-safe border-safe/20' : (rec.tier === 'match' ? 'bg-match/10 text-match border-match/20' : 'bg-reach/10 text-reach border-reach/20')">
                      {{ rec.tier }}
                    </span>
                  </div>

                  <div class="flex items-center gap-6 text-[12px] font-medium text-text-muted mt-4">
                    <div class="flex items-center gap-1.5 bg-bg px-2.5 py-1 rounded-md border border-black/5">
                      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path></svg>
                      {{ $t('caseDetail.likelihood') }}: {{ rec.admission_likelihood_score }}%
                    </div>
                    <div class="flex items-center gap-1.5 bg-bg px-2.5 py-1 rounded-md border border-black/5">
                      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"></path></svg>
                      {{ $t('caseDetail.fit') }}: {{ rec.student_fit_score }}%
                    </div>
                    <div class="flex items-center gap-1.5 bg-bg px-2.5 py-1 rounded-md border border-black/5">
                      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path></svg>
                      {{ $t('caseDetail.rank') }}: #{{ rec.rank_order }}
                    </div>
                  </div>

                </div>
              </div>
            </div>
            
            <div v-else class="flex flex-col items-center justify-center p-12 opacity-50">
               <svg class="w-12 h-12 text-black/20 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"></path></svg>
               <div class="text-[13px] font-medium text-text-muted">{{ $t('caseDetail.underConstruction') }}</div>
            </div>

          </div>
        </div>

      </div>

    </div>
  </div>
</template>
