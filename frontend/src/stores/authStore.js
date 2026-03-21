import { defineStore } from 'pinia'
import { api } from '@/services/api'

function decodeTokenPayload(token) {
  if (!token) return {}
  try {
    const payload = token.split('.')[1]
    const normalized = payload.replace(/-/g, '+').replace(/_/g, '/')
    const decoded = atob(normalized.padEnd(normalized.length + ((4 - normalized.length % 4) % 4), '='))
    return JSON.parse(decoded)
  } catch (error) {
    return {}
  }
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: JSON.parse(localStorage.getItem('user')) || null,
    token: localStorage.getItem('token') || null,
    loading: false,
    error: null
  }),

  getters: {
    isAuthenticated: (state) => !!state.token,
    isAdmin: (state) => state.user?.role === 'admin',
    isVerified: (state) => state.user?.is_verified === true
  },

  actions: {
    setUser(data) {
      const claims = decodeTokenPayload(data.token)
      this.token = data.token
      this.user = {
        id: claims.user_id || null,
        username: data.username,
        role: data.role,
        is_verified: data.is_verified
      }
      localStorage.setItem('token', this.token)
      localStorage.setItem('user', JSON.stringify(this.user))
      
      // Update axios header
      api.defaults.headers.common['Authorization'] = `Bearer ${this.token}`
    },

    logout() {
      this.user = null
      this.token = null
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      delete api.defaults.headers.common['Authorization']
    },

    init() {
      if (this.token) {
        api.defaults.headers.common['Authorization'] = `Bearer ${this.token}`
      }
    }
  }
})
