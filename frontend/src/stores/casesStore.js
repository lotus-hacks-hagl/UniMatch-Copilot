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
      if (filterStatus === 'All' || filterStatus === 'All cases') queryStatus = 'all'
      
      try {
        const params = { status: queryStatus, page: 1, limit: 100 }
        if (assignedTo) {
          params.assigned_to = assignedTo
        }

        const response = await api.get('/cases', { params })
        this.cases = response.data.data || []
        this.total = response.data.meta?.total || 0
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
        this.stats = response.data
      } catch (error) {
        console.error('Failed to fetch stats', error)
      }
    },
    async fetchCasesByDay() {
      try {
        const response = await api.get('/dashboard/cases-by-day')
        return response.data
      } catch (error) {
        console.error('Failed to fetch cases by day', error)
        return []
      }
    },
    async fetchEscalationTrend() {
      try {
        const response = await api.get('/dashboard/escalation-trend')
        return response.data
      } catch (error) {
        console.error('Failed to fetch escalation trend', error)
        return []
      }
    },
    async fetchAnalytics() {
      try {
        const response = await api.get('/dashboard/analytics')
        return response.data
      } catch (error) {
        console.error('Failed to fetch analytics', error)
        return null
      }
    }
  }
})
