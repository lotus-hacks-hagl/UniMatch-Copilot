<script setup>
import { useQueueStore } from '../stores/queueStore'

const queue = useQueueStore()

const items = [
  { id: 'C-2077', student: 'Hoang Thi D', issue: 'AI Confidence Low (< 50%)', sla: '2 hours left', priority: 'High', type: 'Manual override' },
  { id: 'C-2078', student: 'Pham Van E', issue: 'Missing critical transcript', sla: '5 hours left', priority: 'Medium', type: 'Data request' },
  { id: 'C-2079', student: 'Vu Thuy F', issue: 'Budget mismatch with target', sla: '1 day left', priority: 'Low', type: 'Counselor review' },
]
</script>

<template>
  <div class="px-7 py-6 max-w-[1200px] mx-auto space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-xl font-bold text-text flex items-center gap-3">
          Review Queue
          <span class="px-2 py-0.5 rounded-full text-[11px] font-bold bg-reach/10 text-reach border border-reach/20">{{ queue.pendingCount }}</span>
        </h2>
        <p class="text-[13px] text-text-muted mt-1">Cases flagged by TinyFish AI requiring human intervention.</p>
      </div>
      <button class="px-4 py-2 bg-surface border border-black/10 rounded-lg text-[13px] font-medium hover:bg-bg transition-colors">
        Assign to me
      </button>
    </div>

    <div class="bg-surface rounded-xl border border-black/5 shadow-sm overflow-hidden">
      <table class="w-full text-left">
        <thead>
          <tr class="text-[11px] text-text-muted uppercase tracking-wider border-b border-black/5 bg-bg/50">
            <th class="px-5 py-3 font-medium">Case ID</th>
            <th class="px-5 py-3 font-medium">Student</th>
            <th class="px-5 py-3 font-medium">Flag Reason</th>
            <th class="px-5 py-3 font-medium">SLA Deadline</th>
            <th class="px-5 py-3 font-medium">Action</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-black/5 text-[13px]">
          <tr v-for="item in items" :key="item.id" class="hover:bg-bg/50 transition-colors">
            <td class="px-5 py-4 font-medium text-text">{{ item.id }}</td>
            <td class="px-5 py-4 text-text">{{ item.student }}</td>
            <td class="px-5 py-4">
              <div class="flex items-center gap-2 text-reach">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
                <span class="font-medium">{{ item.issue }}</span>
              </div>
              <div class="text-[11px] text-text-muted mt-1">{{ item.type }}</div>
            </td>
            <td class="px-5 py-4">
              <span 
                class="px-2 py-1 rounded text-[11px] font-medium"
                :class="item.priority === 'High' ? 'bg-reach/10 text-reach border border-reach/20' : 'bg-gray-100 text-gray-600'"
              >
                {{ item.sla }}
              </span>
            </td>
            <td class="px-5 py-4">
              <button class="px-3 py-1.5 text-[12px] font-medium text-primary border border-primary/20 rounded-lg hover:bg-secondary transition-colors">
                Resolve
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
