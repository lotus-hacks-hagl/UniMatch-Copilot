import axios from 'axios'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8894/api/v1',
  headers: {
    'Content-Type': 'application/json'
  }
})

// Initialize token if exists
const token = localStorage.getItem('token')
if (token) {
  api.defaults.headers.common['Authorization'] = `Bearer ${token}`
}

// Optional: Interceptors for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Handle unauthorized (e.g. logout)
      localStorage.removeItem('token')
      localStorage.removeItem('user')
    }
    console.error('API Error:', error)
    return Promise.reject(error)
  }
)
