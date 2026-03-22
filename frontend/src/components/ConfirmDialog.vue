<script setup>
import { useConfirm } from '../composables/useConfirm'

const { state, handleConfirm, handleCancel } = useConfirm()
</script>

<template>
  <Transition name="fade">
    <div v-if="state.show" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="handleCancel"></div>
      <div class="card-soft w-full max-w-sm relative z-10 animate-pop-in shadow-[0_20px_60px_rgba(0,0,0,0.2)] rounded-[24px] p-8 text-center">
        <div class="mb-6">
          <div v-if="state.type === 'danger'" class="w-16 h-16 bg-red-50 rounded-full flex items-center justify-center mx-auto mb-4 border border-red-100">
            <svg class="w-8 h-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
          </div>
          <div v-else class="w-16 h-16 bg-[#fff8e1] rounded-full flex items-center justify-center mx-auto mb-4 border border-[#ffecb3]">
            <svg class="w-8 h-8 text-[#f57f17]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
          </div>
          <h3 class="text-xl font-bold text-[#18180f] tracking-tight">{{ state.title }}</h3>
          <p class="text-[14px] text-[#6b6a62] mt-2 leading-relaxed">{{ state.message }}</p>
        </div>
        
        <div class="flex flex-col gap-3">
          <button 
            @click="handleConfirm" 
            class="btn-primary w-full py-3"
            :class="{ 'bg-red-600 hover:bg-red-700': state.type === 'danger' }"
          >
            {{ state.confirmLabel }}
          </button>
          <button @click="handleCancel" class="btn-outline w-full py-3">
            {{ state.cancelLabel }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.animate-pop-in {
  animation: popIn 0.3s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
@keyframes popIn {
  from { opacity: 0; transform: scale(0.92) translateY(10px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}
</style>
