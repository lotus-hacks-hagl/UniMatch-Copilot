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
    normalizeStats(payload) {
      const data = payload?.data || {}
      const avgMinutes = Number(data.avgProcessingTime || 0)
      return {
        casesToday: Number(data.casesToday || 0),
        avgProcessingTime: `${Math.round(avgMinutes)}m`,
        awaitingReview: Number(data.awaitingReview || 0),
        aiConfidenceAvg: Number(data.aiConfidenceAvg || 0),
        activeCrawls: Number(data.activeCrawls || 0)
      }
    },
    async fetchCases(filterStatus = 'All', assignedTo = null, search = '') {
      this.loading = true
      this.filter = filterStatus
      let queryStatus = filterStatus.toLowerCase().replace(' ', '_')
      if (filterStatus === 'All' || filterStatus === 'All cases') queryStatus = 'all'
      
      try {
        const params = { status: queryStatus, page: 1, limit: 100, search }
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
        this.stats = this.normalizeStats(response.data)
      } catch (error) {
        console.error('Failed to fetch stats', error)
      }
    },
    async fetchCasesByDay() {
      try {
        const response = await api.get('/dashboard/cases-by-day')
        return response.data?.data || []
      } catch (error) {
        console.error('Failed to fetch cases by day', error)
        return []
      }
    },
    async fetchEscalationTrend() {
      try {
        const response = await api.get('/dashboard/escalation-trend')
        return response.data?.data || []
      } catch (error) {
        console.error('Failed to fetch escalation trend', error)
        return []
      }
    },
    async fetchAnalytics() {
      try {
        const response = await api.get('/dashboard/analytics')
        return response.data?.data || null
      } catch (error) {
        console.error('Failed to fetch analytics', error)
        return null
      }
    }
  }
})
