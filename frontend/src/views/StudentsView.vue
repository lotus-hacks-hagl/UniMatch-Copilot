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

const isEditing = ref(false)
const editingStudent = ref(null)
const editForm = ref({
  full_name: '',
  intended_major: '',
  background_text: ''
})

const openEditModal = (s) => {
  editingStudent.value = s
  editForm.value = {
    full_name: s.full_name,
    intended_major: s.intended_major,
    background_text: s.background_text || ''
  }
  isEditing.value = true
}

const closeEditModal = () => {
  isEditing.value = false
  editingStudent.value = null
}

const saveStudent = async () => {
  try {
    await api.put(`/students/${editingStudent.value.id}`, editForm.value)
    alert('Student updated successfully!')
    closeEditModal()
    await fetchStudents(pagination.value.page)
  } catch (err) {
    alert('Update failed')
  }
}

const deleteStudent = async (id) => {
  if (!confirm('Are you sure you want to delete this student? This action cannot be undone.')) return
  try {
    await api.delete(`/students/${id}`)
    alert('Student deleted!')
    await fetchStudents(pagination.value.page)
  } catch (err) {
    alert('Delete failed')
  }
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
                  <th class="px-6 py-4 font-bold text-right">Actions</th>
                </tr>
              </thead>
              <TransitionGroup name="list" tag="tbody" class="divide-y divide-black/5 text-[14px]">
                <tr v-for="s in students" :key="s.id" class="hover:bg-gray-50/80 transition-colors group">
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
                  <td class="px-6 py-4 text-right">
                    <div class="flex items-center justify-end gap-2">
                       <button @click="openEditModal(s)" class="p-2 hover:bg-[#a32d2d]/10 rounded-lg text-[#6b6a62] hover:text-[#a32d2d] transition-all">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path></svg>
                       </button>
                       <button @click="deleteStudent(s.id)" class="p-2 hover:bg-red-100 rounded-lg text-[#6b6a62] hover:text-red-600 transition-all">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path></svg>
                       </button>
                    </div>
                  </td>
                </tr>
              </TransitionGroup>
            </table>
          </div>

          <!-- Pagination -->
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

    <!-- Edit Modal -->
    <div v-if="isEditing" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm">
      <div class="bg-white rounded-[24px] shadow-2xl w-full max-w-lg overflow-hidden animate-fade-in">
        <div class="px-8 py-6 border-b border-black/5 flex items-center justify-between bg-[#fafafa]">
          <h3 class="text-lg font-bold text-[#18180f]">Edit Student Profile</h3>
          <button @click="closeEditModal" class="text-[#6b6a62] hover:text-[#18180f]">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>
          </button>
        </div>
        <div class="p-8 space-y-6">
          <div>
            <label class="block text-[13px] font-bold text-[#18180f] mb-2">Full Name</label>
            <input v-model="editForm.full_name" type="text" class="w-full px-4 py-3 rounded-xl border border-black/10 text-[14px] outline-none focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] bg-[#fafafa] focus:bg-white transition-all" />
          </div>
          <div>
            <label class="block text-[13px] font-bold text-[#18180f] mb-2">Desired Major</label>
            <input v-model="editForm.intended_major" type="text" class="w-full px-4 py-3 rounded-xl border border-black/10 text-[14px] outline-none focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] bg-[#fafafa] focus:bg-white transition-all" />
          </div>
          <div>
            <label class="block text-[13px] font-bold text-[#18180f] mb-2">Background Text</label>
            <textarea v-model="editForm.background_text" class="w-full px-4 py-3 rounded-xl border border-black/10 text-[14px] outline-none focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] bg-[#fafafa] focus:bg-white transition-all min-h-[120px] resize-none"></textarea>
          </div>
        </div>
        <div class="px-8 py-6 bg-[#fafafa] border-t border-black/5 flex justify-end gap-3">
          <button @click="closeEditModal" class="btn-outline px-6">Cancel</button>
          <button @click="saveStudent" class="btn-primary px-8 shadow-lg">Save Changes</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.3s ease-out forwards;
}
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px) scale(0.98); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}
</style>
