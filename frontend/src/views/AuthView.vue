<template>
  <div class="min-h-screen bg-[#F5F4F0] flex items-center justify-center p-4 font-sans">
    <div class="w-full max-w-md bg-white border border-[#dfbfbc20] rounded-lg overflow-hidden shadow-sm">
      <!-- Header -->
      <div class="p-8 pb-4 text-center">
        <div class="w-12 h-12 bg-[#a32d2d] rounded flex items-center justify-center mx-auto mb-4">
          <span class="text-white text-2xl font-bold font-serif">U</span>
        </div>
        <h1 class="text-2xl font-bold text-[#18180f] tracking-tight">
          {{ isLogin ? $t('auth.loginTitle') : $t('auth.registerTitle') }}
        </h1>
        <p class="text-[#6b6a62] text-sm mt-1">
          {{ isLogin ? 'Enter your credentials to access your dashboard' : 'Create an account to start managing cases' }}
        </p>
      </div>

      <!-- Form -->
      <div class="p-8 pt-4">
        <form @submit.prevent="handleSubmit" class="space-y-4">
          <div>
            <label class="block text-xs font-semibold text-[#6b6a62] uppercase tracking-wider mb-1">{{ $t('auth.username') }}</label>
            <input 
              v-model="form.username"
              type="text" 
              required
              class="w-full px-4 py-2 bg-[#F5F4F0] border-b-2 border-transparent focus:border-[#a32d2d] focus:outline-none transition-colors text-[#18180f]"
              :placeholder="$t('auth.username')"
            />
          </div>
          <div>
            <label class="block text-xs font-semibold text-[#6b6a62] uppercase tracking-wider mb-1">{{ $t('auth.password') }}</label>
            <input 
              v-model="form.password"
              type="password" 
              required
              class="w-full px-4 py-2 bg-[#F5F4F0] border-b-2 border-transparent focus:border-[#a32d2d] focus:outline-none transition-colors text-[#18180f]"
              placeholder="••••••••"
            />
          </div>

          <div v-if="!isLogin">
            <label class="block text-xs font-semibold text-[#6b6a62] uppercase tracking-wider mb-1">{{ $t('auth.confirmPassword') }}</label>
            <input 
              v-model="form.confirmPassword"
              type="password" 
              required
              class="w-full px-4 py-2 bg-[#F5F4F0] border-b-2 border-transparent focus:border-[#a32d2d] focus:outline-none transition-colors text-[#18180f]"
              placeholder="••••••••"
            />
          </div>

          <div v-if="error" class="text-red-600 text-[13px] font-medium py-1">
            {{ error }}
          </div>

          <button 
            type="submit"
            :disabled="loading"
            class="w-full py-3 bg-gradient-to-br from-[#821419] to-[#a32d2d] text-white font-semibold rounded hover:opacity-90 transition-opacity disabled:opacity-50 mt-4 shadow-sm"
          >
            {{ loading ? $t('auth.processing') : (isLogin ? $t('auth.signIn') : $t('auth.createAccount')) }}
          </button>
        </form>

        <!-- Toggle -->
        <div class="mt-8 text-center text-sm border-t border-[#dfbfbc20] pt-6">
          <span class="text-[#6b6a62]">
            {{ isLogin ? $t('auth.noAccount') : $t('auth.hasAccount') }}
          </span>
          <button 
            @click="isLogin = !isLogin"
            class="ml-1 text-[#a32d2d] font-semibold hover:underline"
          >
            {{ isLogin ? 'Sign up' : 'Log in' }}
          </button>
        </div>
      </div>

      <!-- Footer -->
      <div class="p-4 bg-[#F5F4F0] text-center border-t border-[#dfbfbc20]">
        <p class="text-[10px] text-[#6b6a62] font-medium tracking-widest uppercase">
          Academic Editorial System — UniMatch Copilot v1.0
        </p>
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
