import { defineStore } from 'pinia'

export const useQueueStore = defineStore('queue', {
  state: () => ({
    pendingCount: 3,
    syncCount: 3
  }),
  actions: {
    // In a real app this would fetch from /api/v1/cases/count?status=human_review
    async fetchPendingCount() {
      // Mocked
      this.pendingCount = 3
    }
  }
})
