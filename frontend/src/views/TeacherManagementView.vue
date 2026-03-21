<template>
  <div class="p-8 font-sans">
    <div class="mb-8">
      <h1 class="text-3xl font-bold text-[#18180f] tracking-tight">Teacher Management</h1>
      <p class="text-[#6b6a62] mt-1">Review and verify teacher accounts to grant system access.</p>
    </div>

    <!-- Stats -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
      <div class="bg-white border border-[#dfbfbc40] p-6 rounded-lg">
        <p class="text-xs font-semibold text-[#6b6a62] uppercase tracking-wider mb-2">Total Teachers</p>
        <p class="text-3xl font-bold text-[#18180f] font-serif">{{ teachers.length }}</p>
      </div>
      <div class="bg-white border border-[#dfbfbc40] p-6 rounded-lg">
        <p class="text-xs font-semibold text-[#a32d2d] uppercase tracking-wider mb-2">Pending Verification</p>
        <p class="text-3xl font-bold text-[#a32d2d] font-serif">{{ pendingCount }}</p>
      </div>
      <div class="bg-white border border-[#dfbfbc40] p-6 rounded-lg">
        <p class="text-xs font-semibold text-[#18180f] uppercase tracking-wider mb-2">Verified</p>
        <p class="text-3xl font-bold text-[#18180f] font-serif">{{ verifiedCount }}</p>
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white border border-[#dfbfbc40] rounded-lg overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left">
          <thead>
            <tr class="bg-[#F5F4F0] border-b border-[#dfbfbc40]">
              <th class="px-6 py-4 text-xs font-semibold text-[#6b6a62] uppercase tracking-wider">Teacher</th>
              <th class="px-6 py-4 text-xs font-semibold text-[#6b6a62] uppercase tracking-wider">Registered At</th>
              <th class="px-6 py-4 text-xs font-semibold text-[#6b6a62] uppercase tracking-wider">Status</th>
              <th class="px-6 py-4 text-xs font-semibold text-[#6b6a62] uppercase tracking-wider">Action</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[#dfbfbc20]">
            <tr v-for="teacher in teachers" :key="teacher.id" class="hover:bg-[#F5F4F030] transition-colors">
              <td class="px-6 py-4">
                <div class="flex items-center">
                  <div class="w-8 h-8 bg-[#a32d2d10] rounded flex items-center justify-center mr-3">
                    <span class="text-[#a32d2d] font-bold text-xs">{{ teacher.username.charAt(0).toUpperCase() }}</span>
                  </div>
                  <span class="font-medium text-[#18180f]">{{ teacher.username }}</span>
                </div>
              </td>
              <td class="px-6 py-4 text-sm text-[#6b6a62]">
                {{ new Date(teacher.created_at).toLocaleDateString('en-GB') }}
              </td>
              <td class="px-6 py-4">
                <span 
                  class="px-2 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider"
                  :class="teacher.is_verified ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'"
                >
                  {{ teacher.is_verified ? 'Verified' : 'Pending' }}
                </span>
              </td>
              <td class="px-6 py-4">
                <button 
                  v-if="!teacher.is_verified"
                  @click="handleVerify(teacher.id, true)"
                  class="text-xs font-bold text-[#a32d2d] border border-[#a32d2d] px-3 py-1 rounded hover:bg-[#a32d2d] hover:text-white transition-all"
                >
                  Approve
                </button>
                <button 
                  v-else
                  @click="handleVerify(teacher.id, false)"
                  class="text-xs font-bold text-[#6b6a62] border border-[#6b6a62] px-3 py-1 rounded hover:bg-[#6b6a62] hover:text-white transition-all"
                >
                  Suspend
                </button>
              </td>
            </tr>
            <tr v-if="teachers.length === 0">
              <td colspan="4" class="px-6 py-12 text-center text-[#6b6a62] text-sm italic">
                No teacher accounts found in the system.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { authService } from '../services/authService'

const teachers = ref([])
const loading = ref(false)

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
