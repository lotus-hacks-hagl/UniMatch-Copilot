import { defineStore } from 'pinia'
import { api } from '../services/api'

export const useQueueStore = defineStore('queue', {
  state: () => ({
    pendingCount: 0,
    syncCount: 0
  }),
  actions: {
    async fetchPendingCount() {
      try {
        const response = await api.get('/cases/count?status=human_review')
        // Assume API returns { count: N }
        this.pendingCount = response.data.count || 0
      } catch (error) {
        console.error('Failed to fetch queue count', error)
      }
    },
    async fetchSyncCount() {
      try {
        const response = await api.get('/universities/crawl-active')
        // Assume API returns { count: N }
        this.syncCount = response.data.count || 0
      } catch (error) {
        console.error('Failed to fetch sync count', error)
        this.syncCount = 0
      }
    }
  }
})
