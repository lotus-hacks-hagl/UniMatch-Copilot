<script setup>
import { ref } from 'vue'

const kbs = ref([
  { id: '1', name: 'Massachusetts Institute of Technology (MIT)', location: 'Cambridge, MA', rank: 1, acceptance: '4.8%', tuition: '$57,590' },
  { id: '2', name: 'Stanford University', location: 'Stanford, CA', rank: 2, acceptance: '3.9%', tuition: '$56,169' },
  { id: '3', name: 'Harvard University', location: 'Cambridge, MA', rank: 3, acceptance: '4.0%', tuition: '$54,269' },
  { id: '4', name: 'University of Cambridge', location: 'Cambridge, UK', rank: 4, acceptance: '21.0%', tuition: '£33,825' },
])

const showModal = ref(false)
const formData = ref({
  name: '',
  location: '',
  rank: '',
  acceptance: '',
  tuition: ''
})

const addUniversity = () => {
  kbs.value.push({
    id: Date.now().toString(),
    ...formData.value
  })
  showModal.value = false
  formData.value = { name: '', location: '', rank: '', acceptance: '', tuition: '' }
}
</script>

<template>
  <div class="px-7 py-6 max-w-[1200px] mx-auto space-y-6 relative">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-xl font-bold text-text">University Knowledge Base</h2>
        <p class="text-[13px] text-text-muted mt-1">Explore and filter {{ kbs.length }} synced global universities.</p>
      </div>
      <div class="flex items-center gap-3">
        <div class="relative">
          <input type="text" placeholder="Search univesities..." class="pl-8 pr-3 py-2 text-[13px] bg-surface border border-black/10 focus:border-primary focus:ring-1 focus:ring-primary rounded-lg w-[260px]" />
          <svg class="w-4 h-4 text-text-muted absolute left-2.5 top-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path></svg>
        </div>
        <button class="px-4 py-2 bg-surface text-text border border-black/10 rounded-lg text-[13px] font-medium hover:bg-bg shadow-sm transition-colors">
          A-Z Sort
        </button>
        <button @click="showModal = true" class="px-4 py-2 bg-primary text-white border border-primary rounded-lg text-[13px] font-medium hover:bg-primary-hover shadow-sm transition-colors">
          + Add University
        </button>
      </div>
    </div>

    <div class="bg-surface rounded-xl border border-black/5 shadow-sm overflow-hidden">
      <table class="w-full text-left">
        <thead>
          <tr class="text-[11px] text-text-muted uppercase tracking-wider border-b border-black/5 bg-bg/50">
            <th class="px-5 py-3 font-medium">Global Rank</th>
            <th class="px-5 py-3 font-medium">Institution</th>
            <th class="px-5 py-3 font-medium">Location</th>
            <th class="px-5 py-3 font-medium">Acceptance Rate</th>
            <th class="px-5 py-3 font-medium">Tuition (Est.)</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-black/5 text-[13px]">
          <tr v-for="kb in kbs" :key="kb.id" class="hover:bg-bg/50 transition-colors cursor-pointer group">
            <td class="px-5 py-4 min-w-[100px]">
              <div class="w-8 h-8 rounded-full bg-secondary text-primary font-bold flex items-center justify-center">#{{ kb.rank }}</div>
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
           <h3 class="text-lg font-bold text-text">Add University</h3>
           <button @click="showModal = false" class="text-text-muted hover:text-text transition-colors">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>
           </button>
        </div>
        
        <form @submit.prevent="addUniversity" class="space-y-4">
          <div>
            <label class="block text-[13px] font-medium text-text mb-1">Institution Name</label>
            <input required v-model="formData.name" type="text" class="w-full px-3 py-2 rounded-lg border-black/10 text-[13px] focus:ring-1 focus:ring-primary focus:border-primary bg-bg" placeholder="e.g. Oxford University" />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-[13px] font-medium text-text mb-1">Location</label>
              <input required v-model="formData.location" type="text" class="w-full px-3 py-2 rounded-lg border-black/10 text-[13px] focus:ring-1 focus:ring-primary focus:border-primary bg-bg" placeholder="e.g. Oxford, UK" />
            </div>
            <div>
              <label class="block text-[13px] font-medium text-text mb-1">Global Rank</label>
              <input required v-model="formData.rank" type="number" class="w-full px-3 py-2 rounded-lg border-black/10 text-[13px] focus:ring-1 focus:ring-primary focus:border-primary bg-bg" placeholder="e.g. 5" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-[13px] font-medium text-text mb-1">Acceptance Rate</label>
              <input required v-model="formData.acceptance" type="text" class="w-full px-3 py-2 rounded-lg border-black/10 text-[13px] focus:ring-1 focus:ring-primary focus:border-primary bg-bg" placeholder="e.g. 17.5%" />
            </div>
            <div>
              <label class="block text-[13px] font-medium text-text mb-1">Tuition (Est.)</label>
              <input required v-model="formData.tuition" type="text" class="w-full px-3 py-2 rounded-lg border-black/10 text-[13px] focus:ring-1 focus:ring-primary focus:border-primary bg-bg" placeholder="e.g. £30,000" />
            </div>
          </div>
          
          <div class="pt-4 flex items-center justify-end gap-3 mt-6 border-t border-black/5">
            <button type="button" @click="showModal = false" class="px-4 py-2 text-[13px] font-medium text-text hover:bg-bg rounded-lg transition-colors border border-black/10">Cancel</button>
            <button type="submit" class="px-4 py-2 text-[13px] font-medium text-white bg-primary hover:bg-primary-hover rounded-lg transition-colors shadow-sm">Save University</button>
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
