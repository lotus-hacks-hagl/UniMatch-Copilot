<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../services/api'
import { usePolling } from '../composables/usePolling'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/authStore'
import { useCasesStore } from '../stores/casesStore'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const authStore = useAuthStore()
const casesStore = useCasesStore()

const caseData = ref(null)
const loading = ref(true)
const activeTab = ref('profile')
const editedSummary = ref('')

const tabs = computed(() => {
  const baseTabs = ['profile', 'aiAnalysis', 'documents', 'communication']
  if (caseData.value && !authStore.isAdmin && caseData.value?.assigned_to_id === authStore.user?.id) {
    return [...baseTabs, 'reportEditor']
  }
  return baseTabs
})

const sortedRecommendations = computed(() => {
  if (!caseData.value?.recommendations) return []
  return [...caseData.value.recommendations].sort((a, b) => a.rank_order - b.rank_order)
})

const fetchCase = async () => {
  try {
    const res = await api.get('/cases/' + route.params.id)
    caseData.value = res.data
    editedSummary.value = caseData.value.profile_summary?.main_opinion || ''
  } catch (err) {
    console.error('Fetch case failed', err)
  } finally {
    loading.value = false
  }
}

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

const handleClaim = async () => {
  try {
    await casesStore.claimCase(route.params.id)
    await fetchCase()
    alert('Case claimed successfully!')
  } catch (err) {
    alert('Failed to claim case')
  }
}

const updateReport = async () => {
  try {
    const summary = { ...caseData.value.profile_summary, main_opinion: editedSummary.value }
    await api.put(`/cases/${route.params.id}`, { profile_summary: summary })
    alert('Summary updated!')
    await fetchCase()
  } catch (err) {
    alert('Update failed')
  }
}

const generateReport = async () => {
  try {
    await api.post(`/cases/${route.params.id}/report`)
    alert('Report generation triggered successfully.')
  } catch (err) {
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

      <!-- Layout: Profile Left, Details Right -->
      <div class="flex xl:flex-row flex-col gap-6 items-start">
        
        <!-- Sidebar -->
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
                <p class="text-[13px] text-text-muted">{{ caseData.student?.intended_major }}</p>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-y-5 gap-x-4 mb-6">
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
                <div class="text-[11px] text-text-muted mb-0.5 uppercase tracking-wider">{{ $t('caseDetail.intake') }}</div>
                <div class="text-[14px] font-bold text-text truncate">{{ caseData.student?.target_intake || 'N/A' }}</div>
              </div>
            </div>

            <div class="space-y-3 pt-5 border-t border-black/5">
              <button v-if="!caseData.assigned_to_id && !authStore.isAdmin" 
                      @click="handleClaim" 
                      class="w-full py-2 bg-[#a32d2d] text-white rounded-lg text-[13px] font-bold hover:bg-[#821419] shadow-sm transition-colors">
                Claim this Case
              </button>
              <div v-else-if="caseData.assigned_to" class="text-[12px] text-text-muted bg-gray-50 p-2.5 rounded-lg border border-black/5 flex items-center gap-2">
                <span class="w-1.5 h-1.5 rounded-full bg-safe"></span>
                <span class="font-medium text-text">Assigned to {{ caseData.assigned_to.username }}</span>
              </div>
              
              <button @click="generateReport" class="w-full py-2 bg-primary text-white rounded-lg text-[13px] font-medium hover:bg-primary-hover shadow-sm transition-colors flex items-center justify-center gap-2">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path></svg>
                {{ $t('caseDetail.generateReport') }}
              </button>
            </div>
          </div>
        </div>

        <!-- Main Content Area -->
        <div class="flex-1 w-full bg-surface rounded-xl border border-black/5 shadow-sm min-h-[600px] flex flex-col overflow-hidden">
          
          <div class="flex border-b border-black/5 px-4 pt-2 overflow-x-auto">
            <button 
              v-for="t in tabs" 
              :key="t"
              @click="activeTab = t"
              class="px-5 py-3 text-[13px] font-medium border-b-2 transition-colors whitespace-nowrap"
              :class="activeTab === t ? 'border-primary text-primary' : 'border-transparent text-text-muted hover:text-text'"
            >
              {{ $t('caseDetail.tabs.' + t) }}
            </button>
          </div>

          <div class="p-6 flex-1 bg-bg/20">
            <!-- Profile Tab -->
            <div v-if="activeTab === 'profile'">
              <div class="bg-surface p-6 rounded-xl border border-black/10">
                <h3 class="text-base font-bold text-text mb-4">{{ $t('caseDetail.fullBackground') }}</h3>
                <div class="space-y-6 text-[13px]">
                  <div>
                    <div class="font-medium text-text mb-1 relative pb-2">
                      <span class="absolute bottom-0 left-0 w-8 h-0.5 bg-primary/30"></span>
                      {{ $t('caseDetail.extracurriculars') }}
                    </div>
                    <p class="text-text-muted leading-relaxed whitespace-pre-wrap">{{ caseData.student?.extracurriculars || $t('caseDetail.noneRecorded') }}</p>
                  </div>
                  <div>
                    <div class="font-medium text-text mb-1 relative pb-2">
                      <span class="absolute bottom-0 left-0 w-8 h-0.5 bg-primary/30"></span>
                      {{ $t('caseDetail.achievements') }}
                    </div>
                    <p class="text-text-muted leading-relaxed whitespace-pre-wrap">{{ caseData.student?.achievements || $t('caseDetail.noneRecorded') }}</p>
                  </div>
                </div>
              </div>
            </div>

            <!-- AI Analysis -->
            <div v-if="activeTab === 'aiAnalysis'">
               <div v-if="['pending', 'processing'].includes(caseData.status)" class="flex flex-col items-center justify-center h-[300px]">
                 <div class="w-8 h-8 rounded-full border-2 border-primary border-t-transparent animate-spin mb-4"></div>
                 <div class="text-[13px] font-medium text-primary">{{ $t('caseDetail.aiAnalyzing') }}</div>
               </div>
               <div v-else-if="!caseData.recommendations || caseData.recommendations.length === 0" class="text-center p-10 text-text-muted italic">
                 {{ $t('caseDetail.noRecommendations') }}
               </div>
               <div v-else class="grid grid-cols-1 gap-4">
                  <div v-for="rec in sortedRecommendations" :key="rec.id" class="bg-surface p-5 rounded-xl border border-black/10 hover:border-primary/20 transition-colors shadow-sm relative">
                    <div class="flex justify-between items-start mb-3">
                       <div>
                         <h4 class="font-bold text-text text-base">{{ rec.university_name }}</h4>
                         <p class="text-[12px] text-text-muted mt-1 line-clamp-2">{{ rec.reason }}</p>
                       </div>
                       <span class="px-2.5 py-1 text-[11px] font-bold uppercase tracking-wider rounded border"
                             :class="rec.tier === 'safe' ? 'bg-safe/10 text-safe border-safe/20' : (rec.tier === 'match' ? 'bg-match/10 text-match border-match/20' : 'bg-reach/10 text-reach border-reach/20')">
                         {{ rec.tier }}
                       </span>
                    </div>

                    <div class="flex items-center gap-4 text-[11px] font-medium text-text-muted mt-3">
                      <div class="flex items-center gap-1.5 bg-bg px-2 py-0.5 rounded border border-black/5">
                        {{ $t('caseDetail.likelihood') }}: {{ rec.admission_likelihood_score }}%
                      </div>
                      <div class="flex items-center gap-1.5 bg-bg px-2 py-0.5 rounded border border-black/5">
                        {{ $t('caseDetail.fit') }}: {{ rec.student_fit_score }}%
                      </div>
                      <div class="flex items-center gap-1.5 bg-bg px-2 py-0.5 rounded border border-black/5">
                        {{ $t('caseDetail.rank') }}: #{{ rec.rank_order }}
                      </div>
                    </div>
                  </div>
               </div>
            </div>

            <!-- Report Editor -->
            <div v-if="activeTab === 'reportEditor'">
              <div class="bg-surface rounded-xl p-6 border border-black/10 h-full flex flex-col">
                <div class="flex items-center justify-between mb-4">
                  <h3 class="text-base font-bold text-text">Refine AI Summary</h3>
                  <button @click="updateReport" class="px-4 py-1.5 bg-safe text-white text-[12px] font-bold rounded-lg hover:bg-opacity-90 transition-colors">
                    Save Changes
                  </button>
                </div>
                <p class="text-[12px] text-text-muted mb-4 italic">
                  Personalize the AI's "Main Opinion" here before finalizing the student's report.
                </p>
                <textarea 
                  v-model="editedSummary"
                  class="flex-1 w-full min-h-[300px] p-4 text-[13px] bg-bg rounded-lg border-black/5 focus:border-primary focus:ring-1 focus:ring-primary leading-relaxed"
                ></textarea>
              </div>
            </div>

            <!-- Others -->
            <div v-if="!['profile', 'aiAnalysis', 'reportEditor'].includes(activeTab)" class="text-center p-12 opacity-50">
               <svg class="w-12 h-12 mx-auto text-black/10 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"></path></svg>
               <div class="text-[13px] font-medium text-text-muted">{{ $t('caseDetail.underConstruction') }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
