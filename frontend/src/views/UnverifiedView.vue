<template>
  <div class="min-h-screen bg-[#F5F4F0] flex items-center justify-center p-4">
    <div class="w-full max-w-md bg-white border border-[#dfbfbc40] rounded-lg p-8 text-center">
      <div class="w-16 h-16 bg-[#ba751720] rounded-full flex items-center justify-center mx-auto mb-6">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-[#ba7517]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
        </svg>
      </div>
      <h1 class="text-xl font-bold text-[#18180f] mb-2">Account Pending Verification</h1>
      <p class="text-[#6b6a62] text-sm mb-6">
        Welcome to UniMatch, <strong>{{ authStore.user?.username }}</strong>. Your teacher account has been created successfully but requires manual approval from a Super Admin before you can access student cases.
      </p>
      <div class="bg-[#F5F4F0] p-4 rounded text-xs text-left mb-6 text-[#6b6a62]">
        <p class="font-bold mb-1">What happens next?</p>
        <ul class="list-disc ml-4 space-y-1">
          <li>Admin reviews your application.</li>
          <li>Your status will be updated to "Verified".</li>
          <li>You will be able to claim student cases and generate AI reports.</li>
        </ul>
      </div>
      <button 
        @click="authStore.logout(); router.push('/auth')"
        class="w-full py-2 text-[#a32d2d] font-semibold hover:underline"
      >
        Sign out
      </button>
    </div>
  </div>
</template>

<script setup>
import { useAuthStore } from '../stores/authStore'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()
</script>
