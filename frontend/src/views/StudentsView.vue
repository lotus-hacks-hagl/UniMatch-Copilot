<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../services/api'
import { useI18n } from 'vue-i18n'

const router = useRouter()

const { t } = useI18n()
const students = ref([])
const loading = ref(true)
const pagination = ref({
  page: 1,
  limit: 10,
  total: 0
})

const fetchStudents = async (page = 1) => {
  loading.value = true
  try {
    const res = await api.get('/students', {
      params: { page, limit: pagination.value.limit }
    })
    students.value = res.data.data
    pagination.value = {
      ...pagination.value,
      page: res.data.meta.page,
      total: res.data.meta.total
    }
  } catch (err) {
    console.error('Failed to fetch students:', err)
  } finally {
    loading.value = false
  }
}

onMounted(() => fetchStudents())

const getAvatar = (name) => {
  if (!name) return '??'
  const parts = name.split(' ')
  if (parts.length >= 2) return (parts[0][0] + parts[parts.length-1][0]).toUpperCase()
  return name.substring(0, 2).toUpperCase()
}
</script>

<template>
  <div class="px-8 py-6 max-w-7xl mx-auto space-y-8 font-sans">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-2xl font-bold text-[#18180f] tracking-tight">{{ t('students.title') }}</h2>
        <p class="text-[14px] text-[#6b6a62] mt-1">{{ t('students.subtitle') }}</p>
      </div>
    </div>

    <div class="card-soft overflow-hidden p-0 flex flex-col min-h-[400px]">
      <Transition name="fade" mode="out-in">
        <div v-if="loading" class="flex-1 flex flex-col items-center justify-center p-12 space-y-4">
          <div class="w-10 h-10 border-4 border-red-100 border-t-[#a32d2d] rounded-full animate-spin"></div>
          <div class="text-[14px] font-medium text-[#6b6a62]">{{ t('common.loading') }}</div>
        </div>
        
        <div v-else-if="students.length === 0" class="flex-1 flex flex-col items-center justify-center p-12 text-center">
          <div class="w-20 h-20 bg-gray-50 rounded-full flex items-center justify-center mb-4">
            <svg class="w-10 h-10 text-[#a8a79d]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"></path></svg>
          </div>
          <div class="text-[16px] font-bold text-[#18180f] mb-2">{{ t('students.noData') }}</div>
          <p class="text-[14px] text-[#6b6a62]">Get started by adding a new student profile.</p>
        </div>

        <div v-else class="flex-1 flex flex-col">
          <div class="overflow-x-auto flex-1">
            <table class="w-full text-left border-collapse">
              <thead>
                <tr class="text-[12px] text-[#8a8980] uppercase tracking-wider border-b border-black/5 bg-[#fafafa]">
                  <th class="px-6 py-4 font-bold">{{ t('students.name') }}</th>
                  <th class="px-6 py-4 font-bold">{{ t('students.targetIntake') }}</th>
                  <th class="px-6 py-4 font-bold">{{ t('students.gpa') }}</th>
                  <th class="px-6 py-4 font-bold">{{ t('students.testScores') }}</th>
                  <th class="px-6 py-4 font-bold">{{ t('students.budget') }}</th>
                </tr>
              </thead>
              <TransitionGroup name="list" tag="tbody" class="divide-y divide-black/5 text-[14px]">
                <tr v-for="s in students" :key="s.id" class="hover:bg-gray-50/80 transition-colors group cursor-pointer">
                  <td class="px-6 py-4">
                    <div class="flex items-center gap-3">
                      <div class="w-9 h-9 rounded-full bg-[#f4f5f7] text-[#18180f] font-bold flex items-center justify-center shrink-0 border border-black/5">
                        {{ getAvatar(s.full_name) }}
                      </div>
                      <div>
                        <div class="font-bold text-[#18180f] group-hover:text-[#a32d2d] transition-colors">{{ s.full_name }}</div>
                        <div class="text-[12px] text-[#6b6a62] mt-0.5 line-clamp-1 max-w-[200px]">{{ s.intended_major || 'Undecided Major' }}</div>
                      </div>
                    </div>
                  </td>
                  <td class="px-6 py-4 font-medium text-[#18180f]">{{ s.target_intake || '-' }}</td>
                  <td class="px-6 py-4">
                    <div class="font-bold text-[#18180f]">{{ s.gpa_raw || 0 }} <span class="text-[12px] text-[#8a8980] font-normal">/ {{ s.gpa_scale || '4.0' }}</span></div>
                    <div class="text-[12px] text-[#6b6a62] mt-0.5">Norm: {{ s.gpa_normalized?.toFixed(2) || '0.00' }}</div>
                  </td>
                  <td class="px-6 py-4">
                    <div class="flex flex-wrap gap-2">
                      <span v-if="s.ielts_overall" class="px-2.5 py-1 rounded text-[11px] font-bold bg-[#e8f5e9] text-[#2e7d32]">IELTS {{ s.ielts_overall }}</span>
                      <span v-if="s.sat_total" class="px-2.5 py-1 rounded text-[11px] font-bold bg-[#fff8e1] text-[#f57f17]">SAT {{ s.sat_total }}</span>
                      <span v-if="!s.ielts_overall && !s.sat_total" class="text-[12px] text-[#8a8980] italic">-</span>
                    </div>
                  </td>
                  <td class="px-6 py-4 font-bold text-[#18180f]">
                    ${{ s.budget_usd_per_year?.toLocaleString() || 0 }}/yr
                  </td>
                </tr>
              </TransitionGroup>
            </table>
          </div>

          <!-- Pagination styling update -->
          <div v-if="pagination.total > pagination.limit" class="px-6 py-4 border-t border-black/5 flex items-center justify-between bg-[#fafafa]">
            <span class="text-[13px] font-medium text-[#6b6a62]">
              Showing <span class="text-[#18180f] font-bold">{{ (pagination.page - 1) * pagination.limit + 1 }}</span> to <span class="text-[#18180f] font-bold">{{ Math.min(pagination.page * pagination.limit, pagination.total) }}</span> of <span class="text-[#18180f] font-bold">{{ pagination.total }}</span> students
            </span>
            <div class="flex gap-2">
              <button 
                @click="fetchStudents(pagination.page - 1)" 
                :disabled="pagination.page === 1"
                class="btn-outline px-3 py-1.5 disabled:opacity-50 hover:-translate-y-0"
              >
                Prev
              </button>
              <button 
                @click="fetchStudents(pagination.page + 1)" 
                :disabled="pagination.page * pagination.limit >= pagination.total"
                class="btn-outline px-3 py-1.5 disabled:opacity-50 hover:-translate-y-0"
              >
                Next
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </div>
  </div>
</template>
