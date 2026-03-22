<template>
  <div class="min-h-screen flex font-sans bg-[#f7f8fa] relative overflow-hidden">
    <!-- Left Banner (Red Gradient with Pattern) -->
    <div class="hidden md:block w-[35%] bg-gradient-to-br from-[#7a1315] via-[#a32d2d] to-[#d6c7c7] relative">
      <!-- Decorative academic pattern (SVG) -->
      <div class="absolute inset-0 opacity-10 flex items-center justify-center p-8">
        <svg viewBox="0 0 400 400" fill="none" class="w-full h-auto" stroke="white" stroke-width="1.5">
          <!-- Book -->
          <path d="M50 200 L150 150 L250 200 L150 250 Z" />
          <path d="M50 220 L150 270 L250 220" />
          <path d="M50 240 L150 290 L250 240" />
          <!-- Atom/Science -->
          <ellipse cx="300" cy="100" rx="40" ry="15" transform="rotate(30 300 100)" />
          <ellipse cx="300" cy="100" rx="40" ry="15" transform="rotate(-30 300 100)" />
          <circle cx="300" cy="100" r="5" fill="white" />
          <!-- Graduation Cap -->
          <path d="M100 50 L200 20 L300 50 L200 80 Z" />
          <path d="M150 65 V120 C150 130 250 130 250 120 V65" />
          <path d="M300 50 V100" />
          <!-- Connecting lines and nodes -->
          <circle cx="200" cy="150" r="3" fill="white" />
          <line x1="200" y1="80" x2="200" y2="150" />
          <line x1="200" y1="150" x2="300" y2="100" />
          <line x1="200" y1="150" x2="150" y2="200" />
        </svg>
      </div>
    </div>

    <!-- Right Content Area (Login Card) -->
    <div class="w-full md:w-[65%] flex items-center justify-center p-6 relative">
      <!-- Login Card -->
      <div class="w-full max-w-[440px] bg-white rounded-[24px] shadow-[0_12px_40px_rgba(0,0,0,0.08)] p-10 relative z-10 transition-all duration-500">
        
        <!-- Header -->
        <div class="text-center mb-8">
          <!-- Logo mark -->
          <div class="w-12 h-12 flex items-center justify-center mx-auto mb-4">
            <svg viewBox="0 0 40 40" class="w-10 h-10" fill="none">
              <path d="M8 8 v14 c0 6.6 5.4 12 12 12 s12 -5.4 12 -12 v-14" stroke="#a32d2d" stroke-width="6" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </div>
          <h1 class="text-[28px] font-bold text-[#440b0b] mb-2 tracking-tight">
            {{ isLogin ? $t('auth.loginTitle') : $t('auth.registerTitle') }}
          </h1>
          <p class="text-[13px] text-[#6b6a62] leading-relaxed max-w-[300px] mx-auto">
            {{ isLogin ? 'Enter your registered credentials and we will authenticate your access.' : 'Create your account to start managing student cases.' }}
          </p>
        </div>

        <!-- Form -->
        <form @submit.prevent="handleSubmit" class="space-y-5">
          <div>
            <label class="block text-[11px] font-bold text-[#8a8980] uppercase tracking-widest mb-2">{{ $t('auth.username') }}</label>
            <input 
              v-model="form.username"
              type="text" 
              required
              data-testid="auth-username"
              class="w-full px-4 py-3.5 bg-[#f4f5f7] border border-transparent rounded-xl focus:bg-white focus:border-[#a32d2d] focus:ring-4 focus:ring-[#a32d2d]/10 outline-none transition-all duration-300 text-[#18180f] placeholder-[#a8a79d]"
              placeholder="admin"
            />
          </div>
          <div>
            <label class="block text-[11px] font-bold text-[#8a8980] uppercase tracking-widest mb-2">{{ $t('auth.password') }}</label>
            <input 
              v-model="form.password"
              type="password" 
              required
              data-testid="auth-password"
              class="w-full px-4 py-3.5 bg-[#f4f5f7] border border-transparent rounded-xl focus:bg-white focus:border-[#a32d2d] focus:ring-4 focus:ring-[#a32d2d]/10 outline-none transition-all duration-300 text-[#18180f] placeholder-[#a8a79d]"
              placeholder="admin@123"
            />
          </div>

          <div v-if="!isLogin">
            <label class="block text-[11px] font-bold text-[#8a8980] uppercase tracking-widest mb-2">{{ $t('auth.confirmPassword') }}</label>
            <input 
              v-model="form.confirmPassword"
              type="password" 
              required
              class="w-full px-4 py-3.5 bg-[#f4f5f7] border border-transparent rounded-xl focus:bg-white focus:border-[#a32d2d] focus:ring-4 focus:ring-[#a32d2d]/10 outline-none transition-all duration-300 text-[#18180f] placeholder-[#a8a79d]"
              placeholder="••••••••"
            />
          </div>

          <div v-if="error" class="text-red-600 text-[13px] font-medium py-2 bg-red-50 rounded-lg px-3 border border-red-100 animate-pulse">
            {{ error }}
          </div>

          <!-- Primary Button -->
          <button 
            type="submit"
            :disabled="loading"
            data-testid="auth-submit"
            class="w-full py-4 mt-6 bg-gradient-to-r from-[#8B0000] to-[#b30000] text-white font-bold rounded-full shadow-[0_4px_14px_rgba(163,45,45,0.4)] hover:shadow-[0_6px_20px_rgba(163,45,45,0.6)] hover:-translate-y-1 transition-all duration-300 disabled:opacity-50 disabled:hover:translate-y-0 flex justify-center items-center"
          >
            <span v-if="loading" class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin mr-2"></span>
            {{ loading ? $t('auth.processing') : (isLogin ? $t('auth.signIn') : $t('auth.createAccount')) }}
          </button>
        </form>

        <!-- Toggle -->
        <div class="mt-8 text-center pt-2">
          <button 
            @click="isLogin = !isLogin"
            class="text-[#8B0000] text-[13px] font-bold flex items-center justify-center gap-2 mx-auto hover:-translate-x-1 transition-transform duration-300"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"></path></svg>
            {{ isLogin ? 'Create Account' : 'Back to Login' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/authStore'
import { authService } from '../services/authService'
import { useI18n } from 'vue-i18n'

const router = useRouter()
const authStore = useAuthStore()
const { t } = useI18n()

const isLogin = ref(true)
const loading = ref(false)
const error = ref('')

const form = reactive({
  username: '',
  password: '',
  confirmPassword: ''
})

const handleSubmit = async () => {
  // Client-side validation
  if (form.username.length < 3) {
    error.value = t('auth.errors.usernameTooShort')
    return
  }
  if (form.password.length < 6) {
    error.value = t('auth.errors.passwordTooShort')
    return
  }
  if (!isLogin.value && form.password !== form.confirmPassword) {
    error.value = t('auth.errors.passwordMismatch')
    return
  }

  loading.value = true
  error.value = ''
  
  try {
    let response
    if (isLogin.value) {
      response = await authService.login({
        username: form.username,
        password: form.password
      })
    } else {
      response = await authService.register({
        username: form.username,
        password: form.password
      })
    }
    
    authStore.setUser(response.data)
    
    // Redirect based on role and verification
    if (authStore.isAdmin) {
      router.push('/admin/teachers')
    } else if (!authStore.isVerified) {
      router.push('/unverified')
    } else {
      router.push('/cases')
    }
  } catch (err) {
    // FIX: Parsing nested error structure correctly
    error.value = err.response?.data?.error?.message || err.response?.data?.message || 'Authentication failed'
    console.error('Auth Error:', err.response?.data)
  } finally {
    loading.value = false
  }
}
</script>
