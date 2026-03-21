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
  loading.value = true
  try {
    const res = await api.get('/universities', { params: { page: 1, limit: 100 } })
    const uniList = (res.data.data || []).map(u => ({
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
  <div class="px-8 py-6 max-w-7xl mx-auto space-y-8 font-sans relative">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-2xl font-bold text-[#18180f] tracking-tight">{{ $t('universityKb.title') }}</h2>
        <p class="text-[14px] text-[#6b6a62] mt-1">{{ $t('universityKb.subtitle', { count: totalKbs }) }}</p>
      </div>
      <div class="flex items-center gap-4">
        <div class="relative shadow-sm hover-elevate rounded-lg">
          <input type="text" :placeholder="$t('universityKb.search')" class="pl-9 pr-4 py-2 text-[14px] bg-white border border-black/10 focus:border-[#a32d2d] focus:ring-2 focus:ring-[#a32d2d]/10 outline-none rounded-lg w-[260px] transition-all" />
          <svg class="w-4 h-4 text-[#a8a79d] absolute left-3 top-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path></svg>
        </div>
        <button @click="runSync" class="btn-outline">
          <svg class="w-4 h-4 text-[#6b6a62]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path></svg>
          {{ $t('universityKb.sync') }}
        </button>
        <button @click="showModal = true" class="btn-primary shadow-[0_4px_14px_rgba(163,45,45,0.35)]">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 4v16m8-8H4"></path></svg>
          {{ $t('universityKb.add') }}
        </button>
      </div>
    </div>

    <div class="card-soft overflow-hidden p-0 flex flex-col min-h-[400px]">
      <Transition name="fade" mode="out-in">
        <div v-if="loading" class="flex-1 flex flex-col items-center justify-center p-12 space-y-4">
          <div class="w-10 h-10 border-4 border-red-100 border-t-[#a32d2d] rounded-full animate-spin"></div>
          <div class="text-[14px] font-medium text-[#6b6a62]">{{ $t('universityKb.loading') }}</div>
        </div>
        
        <div v-else-if="kbs.length === 0" class="flex-1 flex flex-col items-center justify-center text-center p-12">
           <div class="w-20 h-20 bg-gray-50 rounded-full flex items-center justify-center mb-4">
            <svg class="w-10 h-10 text-[#a8a79d]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"></path></svg>
          </div>
          <h3 class="text-[16px] font-bold text-[#18180f] mb-1">No universities available</h3>
          <p class="text-[14px] text-[#6b6a62]">Run a sync or manually add your first university institution.</p>
        </div>

        <table v-else class="w-full text-left border-collapse">
          <thead>
            <tr class="text-[12px] text-[#8a8980] uppercase tracking-wider border-b border-black/5 bg-[#fafafa]">
              <th class="px-6 py-4 font-bold">{{ $t('universityKb.table.rank') }}</th>
              <th class="px-6 py-4 font-bold">{{ $t('universityKb.table.institution') }}</th>
              <th class="px-6 py-4 font-bold">{{ $t('universityKb.table.location') }}</th>
              <th class="px-6 py-4 font-bold">{{ $t('universityKb.table.acceptance') }}</th>
              <th class="px-6 py-4 font-bold">{{ $t('universityKb.table.tuition') }}</th>
            </tr>
          </thead>
          <TransitionGroup name="list" tag="tbody" class="divide-y divide-black/5 text-[14px]">
            <tr v-for="kb in kbs" :key="kb.id" class="hover:bg-gray-50/80 transition-colors cursor-pointer group">
              <td class="px-6 py-4 min-w-[100px]">
                <div v-if="kb.rank !== '-'" class="w-10 h-10 rounded-full bg-[#f4f5f7] text-[#18180f] font-bold flex items-center justify-center border border-black/5">#{{ kb.rank }}</div>
                <div v-else class="w-10 h-10 rounded-full bg-gray-50 text-[#8a8980] font-bold flex items-center justify-center border border-black/5">-</div>
              </td>
              <td class="px-6 py-4 font-bold text-[#18180f] group-hover:text-[#a32d2d] transition-colors">{{ kb.name }}</td>
              <td class="px-6 py-4 text-[#6b6a62]">{{ kb.location }}</td>
              <td class="px-6 py-4">
                <span v-if="kb.acceptance !== 'N/A'" class="px-2.5 py-1 rounded-md text-[11px] font-bold bg-[#e8f5e9] text-[#2e7d32]">{{ kb.acceptance }}</span>
                <span v-else class="text-[#8a8980] text-[12px] italic">-</span>
              </td>
              <td class="px-6 py-4 font-bold text-[#18180f]">{{ kb.tuition !== 'N/A' ? kb.tuition + '/yr' : '-' }}</td>
            </tr>
          </TransitionGroup>
        </table>
      </Transition>
    </div>

    <!-- Add University Modal -->
    <Transition name="fade">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showModal = false"></div>
        <div class="card-soft w-full max-w-md relative z-10 animate-fade-in shadow-[0_20px_60px_rgba(0,0,0,0.15)] rounded-[24px] p-8">
          <div class="flex items-center justify-between mb-6">
             <h3 class="text-xl font-bold text-[#18180f]">{{ $t('universityKb.modal.title') }}</h3>
             <button @click="showModal = false" class="text-[#a8a79d] hover:text-[#18180f] transition-colors p-1 bg-gray-50 rounded-full hover:bg-gray-100">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>
             </button>
          </div>
          
          <form @submit.prevent="addUniversity" class="space-y-5">
            <div>
              <label class="block text-[13px] font-bold text-[#18180f] mb-1.5">{{ $t('universityKb.modal.name') }}</label>
              <input required v-model="formData.name" type="text" class="w-full px-4 py-2.5 rounded-lg border border-black/10 text-[14px] focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] outline-none transition-all placeholder-[#a8a79d]" placeholder="e.g. Oxford University" />
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-[13px] font-bold text-[#18180f] mb-1.5">{{ $t('universityKb.modal.location') }}</label>
                <input required v-model="formData.location" type="text" class="w-full px-4 py-2.5 rounded-lg border border-black/10 text-[14px] focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] outline-none transition-all placeholder-[#a8a79d]" placeholder="e.g. Oxford, UK" />
              </div>
              <div>
                <label class="block text-[13px] font-bold text-[#18180f] mb-1.5">{{ $t('universityKb.modal.rank') }}</label>
                <input required v-model="formData.rank" type="number" class="w-full px-4 py-2.5 rounded-lg border border-black/10 text-[14px] focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] outline-none transition-all placeholder-[#a8a79d]" placeholder="e.g. 5" />
              </div>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-[13px] font-bold text-[#18180f] mb-1.5">{{ $t('universityKb.modal.acceptance') }}</label>
                <input required v-model="formData.acceptance" type="text" class="w-full px-4 py-2.5 rounded-lg border border-black/10 text-[14px] focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] outline-none transition-all placeholder-[#a8a79d]" placeholder="e.g. 17.5%" />
              </div>
              <div>
                <label class="block text-[13px] font-bold text-[#18180f] mb-1.5">{{ $t('universityKb.modal.tuition') }}</label>
                <input required v-model="formData.tuition" type="text" class="w-full px-4 py-2.5 rounded-lg border border-black/10 text-[14px] focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] outline-none transition-all placeholder-[#a8a79d]" placeholder="e.g. £30,000" />
              </div>
            </div>
            
            <div class="pt-2 flex items-center justify-end gap-3 mt-6">
              <button type="button" @click="showModal = false" class="btn-outline">Cancel</button>
              <button type="submit" class="btn-primary shadow-[0_4px_14px_rgba(163,45,45,0.35)]">Save University</button>
            </div>
          </form>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.animate-fade-in {
  animation: popIn 0.3s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
@keyframes popIn {
  from { opacity: 0; transform: scale(0.92) translateY(10px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}
</style>
