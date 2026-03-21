import { defineStore } from 'pinia'
import { api } from '../services/api'

export const useCasesStore = defineStore('cases', {
  state: () => ({
    cases: [],
    loading: false,
    filter: 'All', // 'All', 'Done', 'Processing', 'Human review'
    stats: {
      casesToday: 0,
      avgProcessingTime: '',
      awaitingReview: 0,
      aiConfidenceAvg: 0
    },
    total: 0
  }),
  actions: {
    async fetchCases(filterStatus = 'All', assignedTo = null) {
      this.loading = true
      this.filter = filterStatus
      let queryStatus = filterStatus.toLowerCase().replace(' ', '_')
      if (filterStatus === 'All') queryStatus = 'all'
      
      try {
        const params = { status: queryStatus, page: 1, limit: 100 }
        if (assignedTo) {
          params.assigned_to = assignedTo
        }

        const response = await api.get('/cases', { params })
        this.cases = response.data.cases || []
        this.total = response.data.total || 0
      } catch (error) {
        console.error('Failed to fetch cases', error)
        this.cases = []
      } finally {
        this.loading = false
      }
    },
    async claimCase(caseId) {
      try {
        await api.post(`/cases/${caseId}/claim`)
        return true
      } catch (error) {
        console.error('Failed to claim case', error)
        throw error
      }
    },
    async fetchStats() {
      try {
        const response = await api.get('/dashboard/stats')
        if (response.data) {
          this.stats = response.data
        }
      } catch (error) {
        console.error('Failed to fetch stats', error)
      }
    }
  }
})
