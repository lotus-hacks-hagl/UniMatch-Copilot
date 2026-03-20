<script setup>
import { ref } from 'vue'

const metrics = [
  { label: 'Active cases', value: '47', sub: '+3 today', trend: 'up' },
  { label: 'Avg processing', value: '2.4h', sub: '-12% vs last week', trend: 'down' },
  { label: 'Awaiting review', value: '3', sub: 'High priority', trend: 'neutral', isAlert: true },
  { label: 'AI Confidence', value: '92%', sub: '+4% accuracy', trend: 'up' }
]

const filters = ['All cases', 'Done', 'Processing', 'Human review']
const activeFilter = ref('All cases')

const cases = [
  {
    id: 'C-2049',
    student: { name: 'Nguyen Van A', avatar: 'NA', gpa: '3.8', ielts: '7.5' },
    profile: { major: 'Computer Science', country: 'USA', budget: '$40k/yr' },
    target: 'Fall 2026',
    tiers: { safe: 2, match: 4, reach: 1 },
    confidence: 94,
    status: 'Done',
    updated: '10m ago'
  },
  {
    id: 'C-2050',
    student: { name: 'Tran Thi B', avatar: 'TB', gpa: '3.2', ielts: '6.5' },
    profile: { major: 'Business', country: 'UK', budget: '$30k/yr' },
    target: 'Fall 2026',
    tiers: { safe: 3, match: 2, reach: 2 },
    confidence: 88,
    status: 'Processing',
    updated: '1h ago'
  },
  {
    id: 'C-2051',
    student: { name: 'Le Van C', avatar: 'LC', gpa: '3.9', ielts: '8.0' },
    profile: { major: 'Data Science', country: 'Canada', budget: '$45k/yr' },
    target: 'Spring 2027',
    tiers: { safe: 1, match: 5, reach: 3 },
    confidence: 72,
    status: 'Human review',
    updated: '2h ago'
  }
]

const getStatusClass = (status) => {
  if (status === 'Done') return 'bg-safe/10 text-safe border-safe/20'
  if (status === 'Processing') return 'bg-match/10 text-match border-match/20'
  return 'bg-reach/10 text-reach border-reach/20'
}
</script>

<template>
  <div class="px-7 py-6 max-w-[1400px] mx-auto space-y-6">
    
    <!-- Metrics Row -->
    <div class="grid grid-cols-4 gap-4">
      <div 
        v-for="m in metrics" 
        :key="m.label"
        class="bg-surface rounded-xl p-5 border border-black/5 shadow-sm relative overflow-hidden group"
      >
        <div class="absolute inset-y-0 left-0 w-1 bg-primary/0 group-hover:bg-primary transition-colors"></div>
        <div class="text-[13px] text-text-muted mb-1 flex items-center justify-between">
          {{ m.label }}
          <span v-if="m.isAlert" class="w-2 h-2 rounded-full bg-red-500 animate-pulse"></span>
        </div>
        <div class="text-2xl font-bold text-text mb-1">{{ m.value }}</div>
        <div class="text-[11px]" :class="m.trend === 'down' ? 'text-safe' : (m.trend === 'up' ? 'text-text-muted' : 'text-reach')">
          {{ m.sub }}
        </div>
      </div>
    </div>

    <!-- Main Table Section -->
    <div class="bg-surface rounded-xl border border-black/5 shadow-sm overflow-hidden flex flex-col">
      <!-- Tabs & Actions -->
      <div class="border-b border-black/5 px-5 flex items-center justify-between">
        <div class="flex items-center gap-6">
          <button 
            v-for="f in filters" 
            :key="f"
            @click="activeFilter = f"
            class="py-3.5 text-[13px] font-medium border-b-2 transition-colors"
            :class="activeFilter === f ? 'border-primary text-primary' : 'border-transparent text-text-muted hover:text-text'"
          >
            {{ f }}
          </button>
        </div>
        <div class="flex items-center gap-3">
          <div class="relative">
            <input type="text" placeholder="Search cases..." class="pl-8 pr-3 py-1.5 text-[13px] bg-bg border-transparent focus:border-primary focus:ring-1 focus:ring-primary rounded-lg w-[200px]" />
            <svg class="w-4 h-4 text-text-muted absolute left-2.5 top-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path></svg>
          </div>
          <button class="p-1.5 rounded text-text-muted border border-black/10 hover:bg-bg transition-colors">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"></path></svg>
          </button>
        </div>
      </div>

      <!-- Table -->
      <div class="overflow-x-auto min-h-[300px]">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="text-[11px] text-text-muted uppercase tracking-wider border-b border-black/5">
              <th class="px-5 py-3 font-medium">Student</th>
              <th class="px-5 py-3 font-medium">Profile</th>
              <th class="px-5 py-3 font-medium">Target</th>
              <th class="px-5 py-3 font-medium">AI Match</th>
              <th class="px-5 py-3 font-medium cursor-pointer flex items-center gap-1 hover:text-text hover:underline decoration-primary/20 underline-offset-4">Confidence <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg></th>
              <th class="px-5 py-3 font-medium">Status</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-black/5 text-[13px]">
            <tr 
              v-for="c in cases" 
              :key="c.id"
              v-show="activeFilter === 'All cases' || activeFilter === c.status"
              class="hover:bg-bg/50 transition-colors group cursor-pointer"
            >
              <td class="px-5 py-3">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-full bg-secondary text-primary font-medium flex items-center justify-center shrink-0 border border-primary/10">
                    {{ c.student.avatar }}
                  </div>
                  <div>
                    <div class="font-medium text-text group-hover:text-primary transition-colors">{{ c.student.name }}</div>
                    <div class="text-[11px] text-text-muted mt-0.5">GPA {{ c.student.gpa }} • IELTS {{ c.student.ielts }}</div>
                  </div>
                </div>
              </td>
              <td class="px-5 py-3">
                <div class="text-text">{{ c.profile.major }}</div>
                <div class="text-[11px] text-text-muted mt-0.5">{{ c.profile.country }} • {{ c.profile.budget }}</div>
              </td>
              <td class="px-5 py-3">
                <div class="text-text">{{ c.target }}</div>
                <div class="text-[11px] text-text-muted mt-0.5">Bachelor</div>
              </td>
              <td class="px-5 py-3">
                <div class="flex gap-1.5">
                  <span v-if="c.tiers.safe" class="px-2 py-0.5 rounded text-[10px] font-medium bg-safe/10 text-safe border border-safe/20">{{ c.tiers.safe }} Safe</span>
                  <span v-if="c.tiers.match" class="px-2 py-0.5 rounded text-[10px] font-medium bg-match/10 text-match border border-match/20">{{ c.tiers.match }} Match</span>
                  <span v-if="c.tiers.reach" class="px-2 py-0.5 rounded text-[10px] font-medium bg-reach/10 text-reach border border-reach/20">{{ c.tiers.reach }} Reach</span>
                </div>
              </td>
              <td class="px-5 py-3">
                <div class="flex items-center gap-2">
                  <div class="w-16 h-1.5 bg-gray-100 rounded-full overflow-hidden">
                    <div 
                      class="h-full rounded-full" 
                      :class="c.confidence >= 90 ? 'bg-safe' : (c.confidence >= 80 ? 'bg-match' : 'bg-reach')"
                      :style="{ width: c.confidence + '%' }"
                    ></div>
                  </div>
                  <span class="text-[11px] font-medium" :class="c.confidence >= 90 ? 'text-safe' : (c.confidence >= 80 ? 'text-match' : 'text-reach')">{{ c.confidence }}%</span>
                </div>
              </td>
              <td class="px-5 py-3">
                <div class="flex items-center justify-between">
                  <span class="inline-flex items-center px-2 py-0.5 rounded-md text-[11px] font-medium border" :class="getStatusClass(c.status)">
                    {{ c.status }}
                  </span>
                  <span class="text-[10px] text-text-muted opacity-0 group-hover:opacity-100 transition-opacity">{{ c.updated }}</span>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="px-5 py-3 border-t border-black/5 bg-gray-50/50 flex items-center justify-between text-[11px] text-text-muted">
        <div>Showing 3 of 47 cases</div>
        <div class="flex items-center gap-1">
          <button class="p-1 rounded hover:bg-gray-200">Prev</button>
          <span class="px-2">1 / 16</span>
          <button class="p-1 rounded hover:bg-gray-200">Next</button>
        </div>
      </div>
    </div>

    <!-- Analytics Row -->
    <div class="grid grid-cols-3 gap-4">
      <div class="bg-surface rounded-xl p-5 border border-black/5 shadow-sm">
        <div class="text-[13px] text-text-muted mb-4 font-medium">Cases per day</div>
        <div class="h-24 flex items-end gap-2 justify-between">
          <div class="w-full bg-primary/20 hover:bg-primary transition-colors rounded-t" style="height: 40%"></div>
          <div class="w-full bg-primary/20 hover:bg-primary transition-colors rounded-t" style="height: 60%"></div>
          <div class="w-full bg-primary/40 hover:bg-primary transition-colors rounded-t" style="height: 30%"></div>
          <div class="w-full bg-primary/80 hover:bg-primary transition-colors rounded-t" style="height: 90%"></div>
          <div class="w-full bg-primary hover:bg-primary transition-colors rounded-t" style="height: 100%"></div>
          <div class="w-full bg-primary/60 hover:bg-primary transition-colors rounded-t" style="height: 70%"></div>
          <div class="w-full bg-primary/40 hover:bg-primary transition-colors rounded-t" style="height: 50%"></div>
        </div>
      </div>
      <div class="bg-surface rounded-xl p-5 border border-black/5 shadow-sm">
        <div class="text-[13px] text-text-muted mb-4 font-medium flex justify-between">Match tier distribution <span class="text-safe">Safe</span></div>
        <div class="h-24 flex items-center justify-center relative">
          <!-- Fake Donut -->
          <div class="w-20 h-20 rounded-full border-4 border-t-safe border-r-match border-b-reach border-l-gray-100 flex items-center justify-center">
             <span class="text-lg font-bold text-text">142</span>
          </div>
        </div>
      </div>
      <div class="bg-surface rounded-xl p-5 border border-black/5 shadow-sm">
        <div class="text-[13px] text-text-muted mb-4 font-medium">Escalation rate trend</div>
        <div class="h-24 relative overflow-hidden">
           <svg viewBox="0 0 100 40" class="w-full h-full preserve-aspect-ratio cursor-pointer group">
             <path d="M0,35 Q10,30 20,32 T40,20 T60,25 T80,10 T100,5" fill="none" class="stroke-reach opacity-50 group-hover:opacity-100 transition-opacity" stroke-width="2" stroke-linecap="round"/>
             <circle cx="100" cy="5" r="2" class="fill-reach" />
             <path d="M0,40 L0,35 Q10,30 20,32 T40,20 T60,25 T80,10 T100,5 L100,40 Z" class="fill-reach/10" />
           </svg>
        </div>
      </div>
    </div>
  </div>
</template>
