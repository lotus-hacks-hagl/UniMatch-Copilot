<script setup>
import { ref, onMounted } from 'vue'
import { api } from '../services/api'
import { useI18n } from 'vue-i18n'

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
  if (!name) return '?'
  return name.substring(0, 2).toUpperCase()
}
</script>

<template>
  <div class="px-7 py-6 max-w-[1200px] mx-auto space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-xl font-bold text-text">{{ t('students.title') }}</h2>
        <p class="text-[13px] text-text-muted mt-1">{{ t('students.subtitle') }}</p>
      </div>
    </div>

    <div class="bg-surface rounded-xl border border-black/5 shadow-sm overflow-hidden">
      <div v-if="loading" class="p-10 text-center text-text-muted text-[13px]">{{ t('common.loading') }}</div>
      
      <div v-else-if="students.length === 0" class="p-12 text-center">
        <div class="text-text-muted text-[13px] mb-4">{{ t('students.noData') }}</div>
      </div>

      <div v-else class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="border-b border-black/5 bg-bg/50">
              <th class="px-5 py-3 text-[11px] font-bold text-text-muted uppercase tracking-wider">{{ t('students.name') }}</th>
              <th class="px-5 py-3 text-[11px] font-bold text-text-muted uppercase tracking-wider">{{ t('students.targetIntake') }}</th>
              <th class="px-5 py-3 text-[11px] font-bold text-text-muted uppercase tracking-wider">{{ t('students.gpa') }}</th>
              <th class="px-5 py-3 text-[11px] font-bold text-text-muted uppercase tracking-wider">{{ t('students.testScores') }}</th>
              <th class="px-5 py-3 text-[11px] font-bold text-text-muted uppercase tracking-wider">{{ t('students.budget') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-black/5">
            <tr v-for="s in students" :key="s.id" class="hover:bg-bg/30 transition-colors">
              <td class="px-5 py-4">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-full bg-secondary text-primary font-medium text-[11px] flex items-center justify-center border border-primary/10">
                    {{ getAvatar(s.full_name) }}
                  </div>
                  <div>
                    <div class="text-[14px] font-bold text-text">{{ s.full_name }}</div>
                    <div class="text-[11px] text-text-muted line-clamp-1 max-w-[200px]">{{ s.intended_major }}</div>
                  </div>
                </div>
              </td>
              <td class="px-5 py-4 text-[13px] text-text">{{ s.target_intake }}</td>
              <td class="px-5 py-4">
                <div class="text-[13px] font-medium text-text">{{ s.gpa_raw }} / {{ s.gpa_scale }}</div>
                <div class="text-[11px] text-text-muted">Norm: {{ s.gpa_normalized?.toFixed(2) }}</div>
              </td>
              <td class="px-5 py-4">
                <div class="flex flex-wrap gap-2">
                  <span v-if="s.ielts_overall" class="px-1.5 py-0.5 rounded bg-match/10 text-match text-[10px] font-bold border border-match/20 uppercase tracking-tighter">IELTS {{ s.ielts_overall }}</span>
                  <span v-if="s.sat_total" class="px-1.5 py-0.5 rounded bg-primary/10 text-primary text-[10px] font-bold border border-primary/20 uppercase tracking-tighter">SAT {{ s.sat_total }}</span>
                </div>
              </td>
              <td class="px-5 py-4 text-[13px] font-medium text-text">
                ${{ s.budget_usd_per_year?.toLocaleString() }}/yr
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="pagination.total > pagination.limit" class="px-5 py-3 border-t border-black/5 flex items-center justify-between bg-bg/30">
        <span class="text-[12px] text-text-muted">
          Showing {{ (pagination.page - 1) * pagination.limit + 1 }} - {{ Math.min(pagination.page * pagination.limit, pagination.total) }} of {{ pagination.total }} students
        </span>
        <div class="flex gap-2">
          <button 
            @click="fetchStudents(pagination.page - 1)" 
            :disabled="pagination.page === 1"
            class="px-3 py-1 text-[12px] border rounded hover:bg-white disabled:opacity-50"
          >
            Prev
          </button>
          <button 
            @click="fetchStudents(pagination.page + 1)" 
            :disabled="pagination.page * pagination.limit >= pagination.total"
            class="px-3 py-1 text-[12px] border rounded hover:bg-white disabled:opacity-50"
          >
            Next
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
