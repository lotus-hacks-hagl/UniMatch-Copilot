import { api } from './api'

export const authService = {
  async login(credentials) {
    const response = await api.post('/auth/login', credentials)
    return response.data
  },

  async register(credentials) {
    const response = await api.post('/auth/register', credentials)
    return response.data
  },

  async getTeachers() {
    const response = await api.get('/admin/teachers')
    return response.data
  },

  async verifyTeacher(id, isVerified) {
    const response = await api.put(`/admin/teachers/${id}/verify`, { is_verified: isVerified })
    return response.data
  }
}
