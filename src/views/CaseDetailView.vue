<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const activeTab = ref('Profile')
const tabs = ['Profile', 'AI Analysis', 'Documents', 'Communication']

const tiers = [
  { name: 'Safe', count: 2, desc: 'High admission probability based on historical data.' },
  { name: 'Match', count: 4, desc: 'Competitive profile matching university benchmarks.' },
  { name: 'Reach', count: 1, desc: 'Aspirational targets with lower historical success.' }
]
</script>

<template>
  <div class="px-7 py-6 max-w-[1200px] mx-auto space-y-6">
    <!-- Header -->
    <div class="flex items-start justify-between">
      <div class="flex items-center gap-4">
        <button @click="router.push('/cases')" class="p-2 rounded hover:bg-black/5 transition-colors border border-transparent hover:border-black/10 text-text-muted">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"></path></svg>
        </button>
        <div>
          <h2 class="text-xl font-bold text-text flex items-center gap-3">
            Nguyen Van A 
            <span class="px-2 py-0.5 rounded text-[11px] font-medium bg-safe/10 text-safe border border-safe/20">Done</span>
          </h2>
          <p class="text-[13px] text-text-muted mt-1">Applying for Fall 2026 • Computer Science (USA)</p>
        </div>
      </div>
      <div class="flex items-center gap-2">
        <button class="px-4 py-2 bg-surface border border-black/10 rounded-lg text-[13px] font-medium hover:bg-bg transition-colors">
          Download PDF
        </button>
        <button class="px-4 py-2 bg-primary text-white rounded-lg text-[13px] font-medium hover:bg-primary-hover transition-colors shadow-sm">
          Edit Profile
        </button>
      </div>
    </div>

    <!-- AI Verdict Card -->
    <div class="bg-surface rounded-xl border border-primary/20 p-6 flex gap-8 items-center shadow-sm relative overflow-hidden">
      <div class="absolute inset-y-0 right-0 w-[30%] bg-gradient-to-l from-secondary/50 to-transparent pointer-events-none"></div>
      
      <div class="w-24 h-24 rounded-full border-[6px] border-safe flex items-center justify-center shrink-0 bg-safe/5">
        <div class="text-center">
          <div class="text-xl font-bold text-safe">94%</div>
        </div>
      </div>
      <div>
        <h3 class="text-base font-bold text-text mb-1">AI Match Confidence: Excellent</h3>
        <p class="text-[13px] text-text-muted max-w-2xl leading-relaxed">
          The student's GPA (3.8) and IELTS (7.5) strongly align with standard requirements for Computer Science in the USA. TinyFish AI recommends focusing on extracurricular software projects to solidify Match tiers into Safe tiers.
        </p>
      </div>
    </div>

    <!-- Main Content Grid -->
    <div class="grid grid-cols-3 gap-6">
      
      <!-- Left sidebar: Tiers -->
      <div class="space-y-4">
        <h4 class="text-[13px] font-bold text-text uppercase tracking-wider">Target Tiers</h4>
        
        <div v-for="t in tiers" :key="t.name" class="p-4 bg-surface rounded-xl border border-black/5 shadow-sm group hover:border-black/15 transition-colors">
          <div class="flex items-center justify-between mb-2">
            <span class="font-bold text-text">{{ t.name }}</span>
            <span 
              class="w-6 h-6 rounded-full flex items-center justify-center text-[11px] font-bold"
              :class="t.name === 'Safe' ? 'bg-safe/20 text-safe' : (t.name === 'Match' ? 'bg-match/20 text-match' : 'bg-reach/20 text-reach')"
            >{{ t.count }}</span>
          </div>
          <p class="text-[12px] text-text-muted">{{ t.desc }}</p>
        </div>
      </div>

      <!-- Right: Tabbed Details -->
      <div class="col-span-2 bg-surface rounded-xl border border-black/5 overflow-hidden shadow-sm flex flex-col min-h-[500px]">
        <div class="flex border-b border-black/5 px-2 pt-2 bg-bg/50">
          <button 
            v-for="t in tabs" 
            :key="t"
            @click="activeTab = t"
            class="px-5 py-3 text-[13px] font-medium border-b-2 transition-colors"
            :class="activeTab === t ? 'border-primary text-primary bg-surface rounded-t-lg' : 'border-transparent text-text-muted hover:text-text hover:bg-black/5 rounded-t-lg'"
          >
            {{ t }}
          </button>
        </div>
        
        <div class="p-6 flex-1 bg-surface">
          <div v-if="activeTab === 'Profile'" class="animate-fade-in space-y-6">
            <div class="grid grid-cols-2 gap-y-6 gap-x-12 text-[13px]">
              <div>
                <div class="text-text-muted mb-1">Full Name</div>
                <div class="font-medium">Nguyen Van A</div>
              </div>
              <div>
                <div class="text-text-muted mb-1">Email</div>
                <div class="font-medium">nva.student@gmail.com</div>
              </div>
              <div>
                <div class="text-text-muted mb-1">GPA / Scale</div>
                <div class="font-medium">3.8 / 4.0</div>
              </div>
              <div>
                <div class="text-text-muted mb-1">IELTS Score</div>
                <div class="font-medium">7.5 (L:8.0, R:7.5, W:6.5, S:7.0)</div>
              </div>
              <div class="col-span-2">
                <div class="text-text-muted mb-1">Counselor Notes</div>
                <div class="bg-bg p-3 rounded-lg border border-black/5 text-text leading-relaxed">
                  Student is highly motivated and has participated in 2 national algorithmic competitions. 
                  Needs guidance on writing the personal statement to highlight leadership skills.
                </div>
              </div>
            </div>
          </div>
          
          <div v-else class="animate-fade-in flex flex-col items-center justify-center h-full text-text-muted">
            <svg class="w-12 h-12 mb-3 text-black/20" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 002-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path></svg>
            <p>{{ activeTab }} information will be integrated here.</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.animate-fade-in { animation: fadeIn 0.2s ease-out forwards; }
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
</style>
