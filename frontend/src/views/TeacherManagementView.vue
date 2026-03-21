<script setup>
import { ref, computed, onMounted } from 'vue'
import { authService } from '../services/authService'

const teachers = ref([])
const loading = ref(true)

const pendingCount = computed(() => teachers.value.filter(t => !t.is_verified).length)
const verifiedCount = computed(() => teachers.value.filter(t => t.is_verified).length)

const fetchTeachers = async () => {
  loading.value = true
  try {
    const response = await authService.getTeachers()
    teachers.value = response.data
  } catch (err) {
    console.error('Failed to fetch teachers:', err)
  } finally {
    loading.value = false
  }
}

const handleVerify = async (id, status) => {
  try {
    await authService.verifyTeacher(id, status)
    // Update local state
    const index = teachers.value.findIndex(t => t.id === id)
    if (index !== -1) {
      teachers.value[index].is_verified = status
    }
  } catch (err) {
    alert('Failed to update verification status')
  }
}

onMounted(fetchTeachers)
</script>

<template>
  <div class="px-8 py-6 max-w-7xl mx-auto space-y-8 font-sans">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-[#18180f] tracking-tight">Teacher Management</h1>
        <p class="text-[14px] text-[#6b6a62] mt-1">Review and verify teacher accounts to grant system access.</p>
      </div>
      <button @click="fetchTeachers" class="btn-outline">
        <svg class="w-4 h-4 text-[#6b6a62]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path></svg>
        Refresh List
      </button>
    </div>

    <Transition name="fade" mode="out-in">
      <div v-if="loading" class="flex flex-col items-center justify-center p-24 space-y-4">
        <div class="w-10 h-10 border-4 border-red-100 border-t-[#a32d2d] rounded-full animate-spin"></div>
        <div class="text-[14px] font-medium text-[#6b6a62]">Loading teachers...</div>
      </div>
      
      <div v-else class="space-y-8">
        <!-- Stats -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-5">
          <div class="card-soft hover-elevate group">
            <p class="text-[13px] font-bold text-[#6b6a62] uppercase tracking-wider mb-2 group-hover:text-[#18180f] transition-colors">Total Teachers</p>
            <p class="text-[36px] font-bold text-[#18180f] leading-none">{{ teachers.length }}</p>
          </div>
          <div class="card-soft hover-elevate group">
            <p class="text-[13px] font-bold text-[#a32d2d] uppercase tracking-wider mb-2">Pending Verification</p>
            <p class="text-[36px] font-bold text-[#a32d2d] leading-none">{{ pendingCount }}</p>
          </div>
          <div class="card-soft hover-elevate group">
            <p class="text-[13px] font-bold text-[#2e7d32] uppercase tracking-wider mb-2 flex items-center justify-between">Verified <svg v-if="verifiedCount > 0" class="w-5 h-5 opacity-20" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"></path></svg></p>
            <p class="text-[36px] font-bold text-[#18180f] leading-none">{{ verifiedCount }}</p>
          </div>
        </div>

        <!-- Table -->
        <div class="card-soft overflow-hidden p-0 flex flex-col min-h-[300px]">
          <div class="overflow-x-auto">
            <table class="w-full text-left border-collapse">
              <thead>
                <tr class="text-[12px] text-[#8a8980] uppercase tracking-wider border-b border-black/5 bg-[#fafafa]">
                  <th class="px-6 py-4 font-bold">Teacher</th>
                  <th class="px-6 py-4 font-bold">Registered At</th>
                  <th class="px-6 py-4 font-bold">Status</th>
                  <th class="px-6 py-4 font-bold text-right">Action</th>
                </tr>
              </thead>
              <TransitionGroup name="list" tag="tbody" class="divide-y divide-black/5 text-[14px]">
                <tr v-for="teacher in teachers" :key="teacher.id" class="hover:bg-gray-50/80 transition-colors group">
                  <td class="px-6 py-4">
                    <div class="flex items-center gap-3">
                      <div class="w-9 h-9 bg-red-50 rounded-full flex items-center justify-center border border-red-100 shrink-0">
                        <span class="text-[#a32d2d] font-bold text-[13px]">{{ teacher.username.charAt(0).toUpperCase() }}</span>
                      </div>
                      <span class="font-bold text-[#18180f] group-hover:text-[#a32d2d] transition-colors">{{ teacher.username }}</span>
                    </div>
                  </td>
                  <td class="px-6 py-4 font-medium text-[#6b6a62]">
                    {{ new Date(teacher.created_at).toLocaleDateString('en-GB') }}
                  </td>
                  <td class="px-6 py-4 border-l border-black/5 px-6">
                    <span 
                      class="px-2.5 py-1 rounded-md text-[11px] font-bold tracking-wide"
                      :class="teacher.is_verified ? 'bg-[#e8f5e9] text-[#2e7d32]' : 'bg-[#fff8e1] text-[#f57f17]'"
                    >
                      {{ teacher.is_verified ? 'VERIFIED' : 'PENDING' }}
                    </span>
                  </td>
                  <td class="px-6 py-4 text-right">
                    <button 
                      v-if="!teacher.is_verified"
                      @click="handleVerify(teacher.id, true)"
                      class="btn-primary shadow-sm inline-flex hover:-translate-y-0.5 ml-auto"
                    >
                      Approve Access
                    </button>
                    <button 
                      v-else
                      @click="handleVerify(teacher.id, false)"
                      class="btn-outline inline-flex hover:-translate-y-0.5 ml-auto text-red-500 hover:text-red-700 hover:bg-red-50"
                    >
                      Suspend
                    </button>
                  </td>
                </tr>
                <tr v-if="teachers.length === 0">
                  <td colspan="4" class="px-6 py-12 text-center text-[#6b6a62] text-[14px]">
                    <div class="w-16 h-16 bg-gray-50 rounded-full mx-auto flex items-center justify-center mb-3">
                      <svg class="w-8 h-8 text-[#a8a79d]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"></path></svg>
                    </div>
                    No teacher accounts found in the system.
                  </td>
                </tr>
              </TransitionGroup>
            </table>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>
