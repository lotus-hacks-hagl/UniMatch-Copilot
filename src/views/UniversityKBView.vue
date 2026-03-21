<script setup>
import { ref, onMounted, computed } from 'vue'
import { api } from '../services/api'
import { useToast } from '../composables/useToast'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const kbs = ref([])
const totalKbs = ref(0)
const loading = ref(true)
const toast = useToast()

const showModal = ref(false)
const formData = ref({
  name: '',
  location: '',
  rank: '',
  acceptance: '',
  tuition: ''
})

const fetchUniversities = async () => {
  try {
    const res = await api.get('/universities', { params: { page: 1, limit: 100 } })
    // map the API schema: Name, Country, QsRank, AcceptanceRate, TuitionUsdPerYear
    const uniList = (res.data.universities || []).map(u => ({
      id: u.id,
      name: u.name,
      location: u.country || 'N/A',
      rank: u.qs_rank || '-',
      acceptance: u.acceptance_rate ? (u.acceptance_rate * 100).toFixed(1) + '%' : 'N/A',
      tuition: u.tuition_usd_per_year ? `$${u.tuition_usd_per_year}` : 'N/A'
    }))
    kbs.value = uniList
    totalKbs.value = res.data.total || uniList.length
  } catch (error) {
    console.error('Failed to fetch universities', error)
  } finally {
    loading.value = false
  }
}

onMounted(fetchUniversities)

const addUniversity = async () => {
  try {
    await api.post('/universities', {
      name: formData.value.name,
      country: formData.value.location,
      qs_rank: parseInt(formData.value.rank) || null,
      acceptance_rate: formData.value.acceptance ? parseFloat(formData.value.acceptance)/100 : null,
      tuition_usd_per_year: parseInt(formData.value.tuition.replace(/\D/g,'')) || null
    })
    toast.addToast(t('universityKb.toasts.addSuccess'), 'success')
    showModal.value = false
    formData.value = { name: '', location: '', rank: '', acceptance: '', tuition: '' }
    await fetchUniversities()
  } catch(err) {
    toast.addToast(t('universityKb.toasts.addFail'), 'error')
  }
}

const runSync = async () => {
  try {
    await api.post('/universities/crawl-all')
    toast.addToast(t('universityKb.toasts.syncSuccess'), 'success')
  } catch (error) {
    toast.addToast(t('universityKb.toasts.syncFail'), 'error')
  }
}
</script>

<template>
  <div class="px-7 py-6 max-w-[1200px] mx-auto space-y-6 relative">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-xl font-bold text-text">{{ $t('universityKb.title') }}</h2>
        <p class="text-[13px] text-text-muted mt-1">{{ $t('universityKb.subtitle', { count: totalKbs }) }}</p>
      </div>
      <div class="flex items-center gap-3">
        <div class="relative">
          <input type="text" :placeholder="$t('universityKb.search')" class="pl-8 pr-3 py-2 text-[13px] bg-surface border border-black/10 focus:border-primary focus:ring-1 focus:ring-primary rounded-lg w-[260px]" />
          <svg class="w-4 h-4 text-text-muted absolute left-2.5 top-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path></svg>
        </div>
        <button @click="runSync" class="px-4 py-2 bg-surface text-text border border-black/10 rounded-lg text-[13px] font-medium hover:bg-bg shadow-sm transition-colors flex items-center gap-2">
          {{ $t('universityKb.sync') }}
        </button>
        <button @click="showModal = true" class="px-4 py-2 bg-primary text-white border border-primary rounded-lg text-[13px] font-medium hover:bg-primary-hover shadow-sm transition-colors">
          {{ $t('universityKb.add') }}
        </button>
      </div>
    </div>

    <div class="bg-surface rounded-xl border border-black/5 shadow-sm overflow-hidden">
      <div v-if="loading" class="p-10 text-center text-text-muted text-[13px]">{{ $t('universityKb.loading') }}</div>
      <table v-else class="w-full text-left">
        <thead>
          <tr class="text-[11px] text-text-muted uppercase tracking-wider border-b border-black/5 bg-bg/50">
            <th class="px-5 py-3 font-medium">{{ $t('universityKb.table.rank') }}</th>
            <th class="px-5 py-3 font-medium">{{ $t('universityKb.table.institution') }}</th>
            <th class="px-5 py-3 font-medium">{{ $t('universityKb.table.location') }}</th>
            <th class="px-5 py-3 font-medium">{{ $t('universityKb.table.acceptance') }}</th>
            <th class="px-5 py-3 font-medium">{{ $t('universityKb.table.tuition') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-black/5 text-[13px]">
          <tr v-for="kb in kbs" :key="kb.id" class="hover:bg-bg/50 transition-colors cursor-pointer group">
            <td class="px-5 py-4 min-w-[100px]">
              <div v-if="kb.rank !== '-'" class="w-8 h-8 rounded-full bg-secondary text-primary font-bold flex items-center justify-center">#{{ kb.rank }}</div>
              <div v-else class="w-8 h-8 rounded-full bg-gray-100 text-text-muted font-bold flex items-center justify-center">-</div>
            </td>
            <td class="px-5 py-4 font-medium text-text group-hover:text-primary transition-colors">{{ kb.name }}</td>
            <td class="px-5 py-4 text-text-muted">{{ kb.location }}</td>
            <td class="px-5 py-4">
              <span class="px-2 py-1 rounded text-[11px] font-medium bg-safe/10 text-safe border border-safe/20">
                {{ kb.acceptance }}
              </span>
            </td>
            <td class="px-5 py-4 font-mono text-[12px] text-text-muted">{{ kb.tuition }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Add University Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showModal = false"></div>
      <div class="bg-surface w-full max-w-md rounded-xl shadow-xl border border-black/10 relative z-10 p-6 animate-fade-in">
        <div class="flex items-center justify-between mb-5">
           <h3 class="text-lg font-bold text-text">{{ $t('universityKb.modal.title') }}</h3>
           <button @click="showModal = false" class="text-text-muted hover:text-text transition-colors">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>
           </button>
        </div>
        
        <form @submit.prevent="addUniversity" class="space-y-4">
          <div>
            <label class="block text-[13px] font-medium text-text mb-1">{{ $t('universityKb.modal.name') }}</label>
            <input required v-model="formData.name" type="text" class="w-full px-3 py-2 rounded-lg border-black/10 text-[13px] focus:ring-1 focus:ring-primary focus:border-primary bg-bg" placeholder="e.g. Oxford University" />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-[13px] font-medium text-text mb-1">{{ $t('universityKb.modal.location') }}</label>
              <input required v-model="formData.location" type="text" class="w-full px-3 py-2 rounded-lg border-black/10 text-[13px] focus:ring-1 focus:ring-primary focus:border-primary bg-bg" placeholder="e.g. Oxford, UK" />
            </div>
            <div>
              <label class="block text-[13px] font-medium text-text mb-1">{{ $t('universityKb.modal.rank') }}</label>
              <input required v-model="formData.rank" type="number" class="w-full px-3 py-2 rounded-lg border-black/10 text-[13px] focus:ring-1 focus:ring-primary focus:border-primary bg-bg" placeholder="e.g. 5" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-[13px] font-medium text-text mb-1">{{ $t('universityKb.modal.acceptance') }}</label>
              <input required v-model="formData.acceptance" type="text" class="w-full px-3 py-2 rounded-lg border-black/10 text-[13px] focus:ring-1 focus:ring-primary focus:border-primary bg-bg" placeholder="e.g. 17.5%" />
            </div>
            <div>
              <label class="block text-[13px] font-medium text-text mb-1">{{ $t('universityKb.modal.tuition') }}</label>
              <input required v-model="formData.tuition" type="text" class="w-full px-3 py-2 rounded-lg border-black/10 text-[13px] focus:ring-1 focus:ring-primary focus:border-primary bg-bg" placeholder="e.g. £30,000" />
            </div>
          </div>
          
          <div class="pt-4 flex items-center justify-end gap-3 mt-6 border-t border-black/5">
            <button type="button" @click="showModal = false" class="px-4 py-2 text-[13px] font-medium text-text hover:bg-bg rounded-lg transition-colors border border-black/10">{{ $t('universityKb.modal.cancel') }}</button>
            <button type="submit" class="px-4 py-2 text-[13px] font-medium text-white bg-primary hover:bg-primary-hover rounded-lg transition-colors shadow-sm">{{ $t('universityKb.modal.save') }}</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.15s ease-out forwards;
}
@keyframes fadeIn {
  from { opacity: 0; transform: scale(0.95); }
  to { opacity: 1; transform: scale(1); }
}
</style>
