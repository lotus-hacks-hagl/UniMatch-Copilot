<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../services/api'
import { usePolling } from '../composables/usePolling'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/authStore'
import { useCasesStore } from '../stores/casesStore'
import { useConfirm } from '../composables/useConfirm'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const authStore = useAuthStore()
const casesStore = useCasesStore()
const { confirm } = useConfirm()

const caseData = ref(null)
const loading = ref(true)
const activeTab = ref('profile')
const editedSummary = ref('')
const noteText = ref('')
const isAddingNote = ref(false)

const documents = ref([])
const isUploading = ref(false)
const fileInput = ref(null)

const tabs = computed(() => {
  const baseTabs = ['profile', 'aiAnalysis', 'documents', 'communication']
  const isAssignedToCurrentUser = caseData.value && (
    caseData.value?.assigned_to_id === authStore.user?.id ||
    caseData.value?.assigned_to?.username === authStore.user?.username
  )
  if (caseData.value && !authStore.isAdmin && isAssignedToCurrentUser) {
    return [...baseTabs, 'reportEditor']
  }
  return baseTabs
})

const sortedRecommendations = computed(() => {
  if (!caseData.value?.recommendations) return []
  return [...caseData.value.recommendations].sort((a, b) => a.rank_order - b.rank_order)
})

const aiProvenance = computed(() => caseData.value?.profile_summary?.provenance || null)
const aiProvenanceMode = computed(() => aiProvenance.value?.mode || '')
const aiProvenanceNote = computed(() => {
  if (caseData.value?.status === 'human_review' && caseData.value?.escalation_reason) {
    return caseData.value.escalation_reason
  }
  return aiProvenance.value?.note || t('caseDetail.autoApproved')
})
const isHeuristicFallback = computed(() => aiProvenanceMode.value === 'heuristic_fallback')
const isModelFilled = computed(() => ['openai_fill', 'provider_plus_openai_fill'].includes(aiProvenanceMode.value))

const fetchCase = async () => {
  try {
    const res = await api.get('/cases/' + route.params.id)
    if (res.data?.success) {
      caseData.value = res.data.data
      editedSummary.value = caseData.value.profile_summary?.main_opinion || ''
      documents.value = caseData.value.documents || []
    }
  } catch (err) {
    console.error('Fetch case failed', err)
  } finally {
    loading.value = false
  }
}

let pollTimer = null
onMounted(async () => {
  await fetchCase()
  if (caseData.value && ['pending', 'processing'].includes(caseData.value.status)) {
    pollTimer = setInterval(async () => {
      await fetchCase()
      if (!['pending', 'processing'].includes(caseData.value?.status)) {
        clearInterval(pollTimer)
      }
    }, 3000)
  }
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})

const getAvatar = (name) => {
  if (!name) return '??'
  const parts = name.split(' ')
  if (parts.length >= 2) return (parts[0][0] + parts[parts.length-1][0]).toUpperCase()
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

const addNote = async () => {
  if (!noteText.value.trim()) return
  isAddingNote.value = true
  try {
    await api.post(`/cases/${route.params.id}/notes`, { text: noteText.value })
    noteText.value = ''
    await fetchCase()
  } catch (err) {
    alert('Failed to add note')
  } finally {
    isAddingNote.value = false
  }
}

const triggerUpload = () => {
  fileInput.value.click()
}

const onFileSelected = async (event) => {
  const file = event.target.files[0]
  if (!file) return

  const formData = new FormData()
  formData.append('file', file)

  isUploading.value = true
  try {
    await api.post(`/cases/${route.params.id}/documents`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
      await fetchCase()
  } catch (err) {
    alert('Failed to upload document')
  } finally {
    isUploading.value = false
    event.target.value = ''
  }
}

const downloadDoc = (doc) => {
  window.open(`${import.meta.env.VITE_API_BASE_URL || '/api/v1'}/documents/${doc.id}`, '_blank')
}

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const isReAnalyzing = ref(false)
const reAnalyze = async () => {
  const ok = await confirm({
    title: 'Re-analyze Case?',
    message: 'This will clear current recommendations and trigger a new AI analysis based on the latest student profile.',
    type: 'warning',
    confirmLabel: 'Start Analysis'
  })
  if (!ok) return
  isReAnalyzing.value = true
  try {
    await api.post(`/cases/${route.params.id}/analyze`)
    alert('Re-analysis triggered!')
    await fetchCase()
    // Start polling if not already polling
    if (!pollTimer) {
      pollTimer = setInterval(async () => {
        await fetchCase()
        if (!['pending', 'processing'].includes(caseData.value?.status)) {
          clearInterval(pollTimer)
          pollTimer = null
        }
      }, 3000)
    }
  } catch (err) {
    alert('Failed to trigger re-analysis')
  } finally {
    isReAnalyzing.value = false
  }
}

const downloadSummary = () => {
  const student = caseData.value.student
  const summary = caseData.value.profile_summary
  const recs = sortedRecommendations.value

  const printWindow = window.open('', '_blank')
  printWindow.document.write(`
    <html>
      <head>
        <title>Case Summary - ${student?.full_name}</title>
        <style>
          body { font-family: sans-serif; padding: 40px; color: #18180f; line-height: 1.6; }
          .header { border-bottom: 2px solid #a32d2d; padding-bottom: 20px; margin-bottom: 30px; }
          h1 { margin: 0; color: #a32d2d; }
          .section { margin-bottom: 30px; }
          .section-title { font-weight: bold; font-size: 18px; margin-bottom: 10px; border-bottom: 1px solid #eee; }
          .grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
          .label { font-weight: bold; font-size: 12px; color: #6b6a62; text-transform: uppercase; }
          .val { font-size: 15px; }
          .rec-card { border: 1px solid #eee; padding: 15px; border-radius: 10px; margin-bottom: 10px; }
          .tier { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 11px; font-weight: bold; text-transform: uppercase; }
          .tier-safe { background: #e8f5e9; color: #2e7d32; }
          .tier-match { background: #fff8e1; color: #f57f17; }
          .tier-reach { background: #fee2e2; color: #a32d2d; }
        </style>
      </head>
      <body>
        <div class="header">
          <h1>UniMatch Case Summary</h1>
          <p>Generated on ${new Date().toLocaleDateString()}</p>
        </div>
        
        <div class="section">
          <div class="section-title">Student Profile</div>
          <div class="grid">
            <div><div class="label">Full Name</div><div class="val">${student?.full_name}</div></div>
            <div><div class="label">Intended Major</div><div class="val">${student?.intended_major}</div></div>
            <div><div class="label">GPA (Raw/Scale)</div><div class="val">${student?.gpa_raw} / ${student?.gpa_scale}</div></div>
            <div><div class="label">IELTS</div><div class="val">${student?.ielts_overall || 'N/A'}</div></div>
          </div>
        </div>

        <div class="section">
          <div class="section-title">AI Match Analysis</div>
          <p>${summary?.main_opinion || 'No summary available.'}</p>
        </div>

        <div class="section">
          <div class="section-title">University Recommendations</div>
          ${recs.map(r => `
            <div class="rec-card">
              <div style="display: flex; justify-content: space-between;">
                <span class="val" style="font-weight:bold">${r.university_name}</span>
                <span class="tier tier-${r.tier}">${r.tier}</span>
              </div>
              <p style="font-size: 13px; color: #6b6a62; margin: 8px 0;">${r.reason}</p>
              <div style="font-size: 11px; color: #8a8980;">Likelihood: ${r.admission_likelihood_score}% | Fit: ${r.student_fit_score}%</div>
            </div>
          `).join('')}
        </div>

        <script>
          window.onload = function() { window.print(); window.close(); }
        </script>
      </body>
    </html>
  `)
  printWindow.document.close()
}
</script>

<template>
  <div class="px-8 py-6 max-w-7xl mx-auto space-y-8 font-sans">
    
    <Transition name="fade" mode="out-in">
      <div v-if="loading" class="flex flex-col items-center justify-center p-24 space-y-4">
        <div class="w-10 h-10 border-4 border-red-100 border-t-[#a32d2d] rounded-full animate-spin"></div>
        <div class="text-[14px] font-medium text-[#6b6a62]">{{ $t('caseDetail.loading') }}</div>
      </div>
      
      <div v-else-if="caseData" class="space-y-8">
        
        <!-- Meta & Breadcrumb -->
        <div class="flex items-center gap-3 text-[13px]">
          <router-link to="/cases" class="text-[#6b6a62] hover:text-[#18180f] flex items-center gap-1.5 transition-colors font-medium">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M10 19l-7-7m0 0l7-7m-7 7h18"></path></svg>
            {{ $t('common.cases') }}
          </router-link>
          <span class="text-black/20">/</span>
          <span class="font-bold text-[#18180f]">{{ caseData.student?.full_name || $t('caseDetail.student') }}</span>
          
          <span class="ml-4 px-2.5 py-1 rounded-md text-[11px] font-bold border uppercase tracking-wider"
                :class="{
                  'bg-[#e8f5e9] text-[#2e7d32] border-[#c8e6c9]': caseData.status === 'done',
                  'bg-[#fff8e1] text-[#f57f17] border-[#ffecb3]': caseData.status === 'processing',
                  'bg-red-50 text-[#a32d2d] border-red-100': caseData.status === 'human_review',
                  'bg-[#f4f5f7] text-[#6b6a62] border-black/10': !['done','processing','human_review'].includes(caseData.status)
                }">
            {{ (caseData.status || 'pending').replace('_', ' ') }}
          </span>
        </div>

        <!-- Layout: Profile Left, Details Right -->
        <div class="flex xl:flex-row flex-col gap-6 items-start">
          
          <!-- Sidebar -->
          <div class="w-full xl:w-[320px] shrink-0 space-y-5">
            
            <!-- Confidence Card -->
            <div class="card-soft hover-elevate relative overflow-hidden group">
              <div class="absolute inset-x-0 top-0 h-1.5" 
                   :class="formatConfidence(caseData.ai_confidence) >= 90 ? 'bg-gradient-to-r from-[#4caf50] to-[#2e7d32]' : (formatConfidence(caseData.ai_confidence)>=80 ? 'bg-gradient-to-r from-[#ffca28] to-[#f57f17]' : 'bg-gradient-to-r from-[#ce3e3e] to-[#8B0000]')"></div>
              <div class="flex items-start justify-between mb-2 mt-1">
                <div class="text-[13px] font-bold text-[#6b6a62] group-hover:text-[#18180f] transition-colors mt-1">{{ $t('caseDetail.aiMatchVerdict') }}</div>
                <div class="text-[36px] font-bold leading-none" :class="formatConfidence(caseData.ai_confidence)>=90?'text-[#2e7d32]':(formatConfidence(caseData.ai_confidence)>=80?'text-[#f57f17]':'text-[#a32d2d]')">
                  {{ formatConfidence(caseData.ai_confidence) }}%
                </div>
              </div>
              <div v-if="aiProvenanceMode" class="mb-2">
                <span class="inline-flex items-center px-2.5 py-1 rounded-md text-[11px] font-bold border uppercase tracking-wider"
                      :class="isHeuristicFallback ? 'bg-red-50 text-[#a32d2d] border-red-100' : (isModelFilled ? 'bg-[#fff8e1] text-[#f57f17] border-[#ffecb3]' : 'bg-[#e8f5e9] text-[#2e7d32] border-[#c8e6c9]')">
                  {{ aiProvenanceMode.replaceAll('_', ' ') }}
                </span>
              </div>
              <p class="text-[13px] text-[#6b6a62] leading-relaxed">
                {{ aiProvenanceNote }}
              </p>
            </div>

            <!-- Student Card -->
            <div class="card-soft hover-elevate cursor-default group">
              <div class="flex items-center gap-4 border-b border-black/5 pb-5 mb-5">
                <div class="w-14 h-14 rounded-full bg-[#f4f5f7] border border-black/5 text-[#18180f] text-xl font-bold flex items-center justify-center shrink-0">
                  {{ getAvatar(caseData.student?.full_name) }}
                </div>
                <div>
                  <h2 class="text-xl font-bold text-[#18180f] mb-0.5 group-hover:text-[#a32d2d] transition-colors">{{ caseData.student?.full_name }}</h2>
                  <p class="text-[13px] font-medium text-[#6b6a62]">{{ caseData.student?.intended_major }}</p>
                </div>
              </div>

              <div class="grid grid-cols-2 gap-y-6 gap-x-4 mb-6">
                <div>
                  <div class="text-[11px] text-[#8a8980] font-bold mb-1 uppercase tracking-wider">{{ $t('caseDetail.gpaNorm') }}</div>
                  <div class="text-[15px] font-bold text-[#18180f]">{{ caseData.student?.gpa_normalized || 'N/A' }}</div>
                </div>
                <div>
                  <div class="text-[11px] text-[#8a8980] font-bold mb-1 uppercase tracking-wider">{{ $t('caseDetail.ielts') }}</div>
                  <div class="text-[15px] font-bold text-[#18180f]">{{ caseData.student?.ielts_overall || 'N/A' }}</div>
                </div>
                <div>
                  <div class="text-[11px] text-[#8a8980] font-bold mb-1 uppercase tracking-wider">{{ $t('caseDetail.budget') }}</div>
                  <div class="text-[15px] font-bold text-[#18180f]">{{ formatBudget(caseData.student?.budget_usd_per_year) }}</div>
                </div>
                <div>
                  <div class="text-[11px] text-[#8a8980] font-bold mb-1 uppercase tracking-wider">{{ $t('caseDetail.intake') }}</div>
                  <div class="text-[15px] font-bold text-[#18180f] truncate">{{ caseData.student?.target_intake || 'N/A' }}</div>
                </div>
              </div>

              <div class="space-y-3 pt-5 border-t border-black/5">
                <button v-if="!caseData.assigned_to_id && !authStore.isAdmin" 
                        @click="handleClaim" 
                        data-testid="case-claim"
                        class="btn-primary w-full shadow-[0_4px_14px_rgba(163,45,45,0.35)]">
                  Claim this Case
                </button>
                <div v-else-if="caseData.assigned_to" class="text-[13px] text-[#6b6a62] bg-[#f4f5f7] p-3 rounded-xl border border-black/5 flex items-center justify-center gap-2">
                  <span class="w-2 h-2 rounded-full bg-[#2e7d32] animate-pulse"></span>
                  <span class="font-bold text-[#18180f]">Assigned to {{ caseData.assigned_to.username }}</span>
                </div>
                <button @click="generateReport" class="btn-outline w-full hover:-translate-y-0.5">
                  {{ $t('caseDetail.generateReport') }}
                </button>
                <button @click="downloadSummary" class="btn-outline w-full hover:-translate-y-0.5">
                  <svg class="w-4 h-4 text-[#6b6a62]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"></path></svg>
                  Download Summary PDF
                </button>
              </div>
            </div>
          </div>

          <!-- Main Content Area -->
          <div class="flex-1 w-full card-soft p-0 min-h-[600px] flex flex-col">
            
            <div class="flex border-b border-black/5 px-6 pt-2 overflow-x-auto">
                <button 
                  v-for="tab in tabs" 
                  :key="tab"
                  @click="activeTab = tab"
                  :data-testid="`case-tab-${tab}`"
                  class="px-5 py-4 text-[14px] font-bold border-b-[3px] transition-colors whitespace-nowrap"
                  :class="activeTab === tab ? 'border-[#a32d2d] text-[#a32d2d]' : 'border-transparent text-[#6b6a62] hover:text-[#18180f]'"
                >
                {{ $t('caseDetail.tabs.' + tab) }}
              </button>
            </div>

            <div class="p-8 flex-1 bg-[#fafafa]/50">
              <Transition name="fade" mode="out-in">
                
                <!-- Profile Tab -->
                <div v-if="activeTab === 'profile'" key="profile">
                  <div class="bg-white p-8 rounded-[20px] shadow-sm border border-black/5">
                    <h3 class="text-lg font-bold text-[#18180f] mb-6">{{ $t('caseDetail.fullBackground') }}</h3>
                    <div class="space-y-8 text-[14px]">
                      <div>
                        <div class="font-bold text-[#18180f] mb-3 relative inline-block">
                          {{ $t('caseDetail.extracurriculars') }}
                          <span class="absolute -bottom-1 left-0 w-12 h-1 bg-[#a32d2d] rounded-full"></span>
                        </div>
                        <p class="text-[#6b6a62] leading-relaxed whitespace-pre-wrap mt-2">{{ caseData.student?.extracurriculars || $t('caseDetail.noneRecorded') }}</p>
                      </div>
                      <div class="pt-4 border-t border-black/5">
                        <div class="font-bold text-[#18180f] mb-3 relative inline-block">
                          {{ $t('caseDetail.achievements') }}
                          <span class="absolute -bottom-1 left-0 w-12 h-1 bg-[#a32d2d] rounded-full"></span>
                        </div>
                        <p class="text-[#6b6a62] leading-relaxed whitespace-pre-wrap mt-2">{{ caseData.student?.achievements || $t('caseDetail.noneRecorded') }}</p>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- AI Analysis -->
                <div v-else-if="activeTab === 'aiAnalysis'" key="aiAnalysis">
                   <div v-if="['pending', 'processing'].includes(caseData.status)" class="flex flex-col items-center justify-center p-24 text-center">
                     <div class="w-12 h-12 rounded-full border-4 border-red-100 border-t-[#a32d2d] animate-spin mb-4"></div>
                     <div class="text-[15px] font-bold text-[#a32d2d]">{{ $t('caseDetail.aiAnalyzing') }}</div>
                     <p class="text-[13px] text-[#6b6a62] mt-2">The Copilot is matching profiles against the global knowledge base...</p>
                   </div>
                <div v-else-if="!caseData.recommendations || caseData.recommendations.length === 0" class="flex flex-col items-center justify-center p-24 text-center">
                     <div class="w-20 h-20 bg-gray-50 rounded-full flex items-center justify-center mb-4">
                       <svg class="w-10 h-10 text-[#a8a79d]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"></path></svg>
                     </div>
                      <h3 class="text-[16px] font-bold text-[#18180f] mb-1">No AI recommendations found</h3>
                      <p class="text-[14px] text-[#6b6a62] mb-6">Trigger a re-analysis or manually edit the target list.</p>
                      <button @click="reAnalyze" :disabled="isReAnalyzing" class="btn-primary px-8">
                        {{ isReAnalyzing ? 'Starting AI...' : 'Run Analysis Now' }}
                      </button>
                    </div>
                   <TransitionGroup v-else name="list" tag="div" class="grid grid-cols-1 gap-5">
                      <div v-if="aiProvenanceMode && aiProvenanceMode !== 'provider_backed'" class="bg-[#fff8e1] border border-[#ffecb3] rounded-[16px] p-4 text-[13px] text-[#7a5610]">
                        <div class="font-bold mb-1">AI provenance</div>
                        <div>{{ aiProvenanceNote }}</div>
                      </div>
                      <div v-for="rec in sortedRecommendations" :key="rec.id" class="bg-white p-6 rounded-[20px] shadow-sm border border-black/5 hover:shadow-md hover:-translate-y-1 transition-all duration-300 relative group cursor-pointer">
                        <div class="flex justify-between items-start mb-4">
                           <div>
                             <h4 class="font-bold text-[#18180f] text-[16px] group-hover:text-[#a32d2d] transition-colors">{{ rec.university_name }}</h4>
                             <p class="text-[13px] font-medium text-[#6b6a62] mt-1.5 line-clamp-2 leading-relaxed">{{ rec.reason }}</p>
                           </div>
                           <span class="px-2.5 py-1 text-[11px] font-bold uppercase tracking-wider rounded-md border shrink-0 mt-0.5 ml-4"
                                 :class="rec.tier === 'safe' ? 'bg-[#e8f5e9] text-[#2e7d32] border-[#c8e6c9]' : (rec.tier === 'match' ? 'bg-[#fff8e1] text-[#f57f17] border-[#ffecb3]' : 'bg-red-50 text-[#a32d2d] border-red-100')">
                             {{ rec.tier }}
                           </span>
                        </div>

                        <div class="flex items-center gap-4 text-[12px] font-bold text-[#6b6a62] mt-4 pt-4 border-t border-black/5">
                          <div class="flex items-center gap-1.5 bg-[#f4f5f7] px-3 py-1.5 rounded-lg border border-black/5">
                            {{ $t('caseDetail.likelihood') }} <span class="text-[#18180f] ml-1">{{ rec.admission_likelihood_score }}%</span>
                          </div>
                          <div class="flex items-center gap-1.5 bg-[#f4f5f7] px-3 py-1.5 rounded-lg border border-black/5">
                            {{ $t('caseDetail.fit') }} <span class="text-[#18180f] ml-1">{{ rec.student_fit_score }}%</span>
                          </div>
                          <div class="flex items-center gap-1.5 bg-red-50 text-[#a32d2d] border-red-100 px-3 py-1.5 rounded-lg border">
                            {{ $t('caseDetail.rank') }} <span class="ml-1">#{{ rec.rank_order }}</span>
                          </div>
                        </div>
                      </div>
                   </TransitionGroup>
                </div>

                <!-- Report Editor -->
                <div v-else-if="activeTab === 'reportEditor'" key="reportEditor">
                  <div class="bg-white rounded-[20px] p-8 border border-black/5 shadow-sm h-full flex flex-col">
                    <div class="flex items-center justify-between mb-4">
                      <h3 class="text-lg font-bold text-[#18180f]">Refine AI Summary</h3>
                      <button @click="updateReport" data-testid="case-save-report-summary" class="btn-primary shadow-[0_4px_14px_rgba(163,45,45,0.35)]">
                        Save Changes
                      </button>
                    </div>
                    <p class="text-[14px] text-[#6b6a62] mb-6">
                      Personalize the AI's "Main Opinion" here before finalizing the student's report.
                    </p>
                    <textarea 
                      v-model="editedSummary"
                      data-testid="case-report-summary"
                      class="flex-1 w-full min-h-[400px] p-5 text-[14px] bg-[#fafafa] rounded-xl border border-black/5 focus:border-[#a32d2d] focus:ring-2 focus:ring-[#a32d2d]/10 focus:bg-white outline-none transition-all leading-relaxed resize-y"
                    ></textarea>
                  </div>
                </div>

                <!-- Documents Tab -->
                <div v-else-if="activeTab === 'documents'" key="documents" class="space-y-6">
                  <div class="bg-white p-8 rounded-[20px] shadow-sm border border-black/5">
                    <div class="flex items-center justify-between mb-6">
                      <h3 class="text-lg font-bold text-[#18180f]">Contract & Documents</h3>
                      <div class="flex items-center gap-3">
                        <input type="file" ref="fileInput" class="hidden" @change="onFileSelected" />
                        <button @click="triggerUpload" :disabled="isUploading" class="btn-secondary px-4 py-2 flex items-center gap-2">
                          <svg v-if="!isUploading" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a2 2 0 002 2h12a2 2 0 002-2v-1m-4-8l-4-4m0 0L8 8m4-4v12"></path></svg>
                          <div v-else class="w-4 h-4 border-2 border-black/10 border-t-[#a32d2d] rounded-full animate-spin"></div>
                          {{ isUploading ? 'Uploading...' : 'Upload File' }}
                        </button>
                      </div>
                    </div>
                    
                    <div v-if="documents.length > 0" class="space-y-4">
                      <div v-for="doc in documents" :key="doc.id" class="p-4 bg-[#f4f5f7] rounded-xl border border-black/5 flex items-center justify-between group hover:bg-white hover:border-[#a32d2d]/30 transition-all duration-300">
                        <div class="flex items-center gap-3">
                          <div class="w-10 h-10 bg-white rounded-lg flex items-center justify-center shadow-sm font-bold text-[10px] uppercase overflow-hidden"
                               :class="doc.file_type.includes('pdf') ? 'text-red-600' : 'text-blue-600'">
                            {{ doc.file_name.split('.').pop() }}
                          </div>
                          <div>
                            <div class="text-[14px] font-bold text-[#18180f] group-hover:text-[#a32d2d] transition-colors">{{ doc.file_name }}</div>
                            <div class="text-[12px] text-[#6b6a62]">{{ new Date(doc.created_at).toLocaleDateString() }} • {{ formatFileSize(doc.file_size) }}</div>
                          </div>
                        </div>
                        <button @click="downloadDoc(doc)" class="text-[#a32d2d] font-bold text-[13px] hover:underline flex items-center gap-1.5 p-2 bg-white rounded-lg shadow-sm border border-black/5 opacity-0 group-hover:opacity-100 transition-all">
                          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a2 2 0 002 2h12a2 2 0 002-2v-1m-4-4l-4 4m0 0L8 12m4 4V4"></path></svg>
                          Download
                        </button>
                      </div>
                    </div>
                    <div v-else class="text-center py-12 bg-[#fafafa] rounded-2xl border border-dashed border-black/10">
                      <div class="w-12 h-12 bg-gray-50 rounded-full flex items-center justify-center mx-auto mb-3">
                        <svg class="w-6 h-6 text-[#a8a79d]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path></svg>
                      </div>
                      <div class="text-[14px] font-medium text-[#6b6a62]">No documents uploaded yet</div>
                    </div>
                  </div>
                </div>

                <!-- Communication Tab -->
                <div v-else-if="activeTab === 'communication'" key="communication" class="space-y-6">
                  <div class="bg-white p-8 rounded-[20px] shadow-sm border border-black/5 flex flex-col min-h-[500px]">
                    <h3 class="text-lg font-bold text-[#18180f] mb-6">Internal Notes & Activity</h3>
                    
                    <!-- Notes Feed -->
                    <div class="flex-1 space-y-6 mb-8 overflow-y-auto max-h-[400px] pr-2">
                      <div v-for="log in (caseData.activity_logs || []).slice().reverse()" :key="log.id" class="flex gap-4">
                        <div class="w-10 h-10 rounded-full bg-[#f4f5f7] border border-black/5 flex items-center justify-center shrink-0 font-bold text-[13px]">
                          {{ getAvatar(log.user?.username || 'System') }}
                        </div>
                        <div class="flex-1">
                          <div class="flex items-center gap-2 mb-1">
                            <span class="font-bold text-[14px] text-[#18180f]">{{ log.user?.username || 'System' }}</span>
                            <span class="text-[11px] text-[#6b6a62]">{{ new Date(log.created_at).toLocaleString() }}</span>
                          </div>
                          <div class="text-[14px] text-[#6b6a62] p-4 bg-[#fafafa] rounded-2xl rounded-tl-none border border-black/5 leading-relaxed">
                            {{ log.event_type === 'case_note' ? log.details : (log.event_type.replaceAll('_', ' ') + ': ' + log.details) }}
                          </div>
                        </div>
                      </div>
                      <div v-if="!caseData.activity_logs?.length" class="text-center py-20 opacity-50">
                        <p class="text-[14px]">No activity recorded yet.</p>
                      </div>
                    </div>

                    <!-- Input -->
                    <div class="relative">
                      <textarea 
                        v-model="noteText"
                        placeholder="Type a note or update..." 
                        class="w-full p-4 pr-32 bg-[#f4f5f7] rounded-xl border border-black/5 focus:bg-white focus:border-[#a32d2d] focus:ring-2 focus:ring-[#a32d2d]/10 outline-none transition-all text-[14px] min-h-[100px] resize-none"
                      ></textarea>
                      <button 
                        @click="addNote"
                        :disabled="!noteText.trim() || isAddingNote"
                        class="absolute bottom-4 right-4 btn-primary px-6 h-10 flex items-center justify-center shadow-lg disabled:opacity-50"
                      >
                        {{ isAddingNote ? 'Saving...' : 'Post Note' }}
                      </button>
                    </div>
                  </div>
                </div>

                <!-- Others (Fallback) -->
                <div v-else key="others" class="flex flex-col items-center justify-center p-24 text-center opacity-70">
                   <div class="w-20 h-20 bg-gray-50 rounded-full flex items-center justify-center mb-4">
                     <svg class="w-10 h-10 text-[#a8a79d]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"></path></svg>
                   </div>
                   <div class="text-[16px] font-bold text-[#18180f] mb-1">Under Construction</div>
                   <p class="text-[14px] text-[#6b6a62]">This module is currently being built.</p>
                </div>
              </Transition>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>
