<script setup>
import { useToast } from '../composables/useToast'

const { toasts } = useToast()

const toneClass = (type) => {
  if (type === 'success') return 'border-[#c8e6c9] bg-[#e8f5e9] text-[#2e7d32]'
  if (type === 'error') return 'border-red-100 bg-red-50 text-[#a32d2d]'
  return 'border-black/5 bg-white text-[#18180f]'
}
</script>

<template>
  <div class="fixed right-6 top-6 z-[120] flex w-[340px] max-w-[calc(100vw-2rem)] flex-col gap-3 pointer-events-none">
    <TransitionGroup name="toast" tag="div" class="space-y-3">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        :data-testid="`toast-${toast.type}`"
        class="pointer-events-auto rounded-2xl border px-4 py-3 shadow-[0_10px_24px_rgba(0,0,0,0.08)] backdrop-blur-sm"
        :class="toneClass(toast.type)"
      >
        <div class="text-[12px] font-bold uppercase tracking-wider opacity-70">{{ toast.type }}</div>
        <div class="mt-1 text-[14px] font-medium leading-relaxed">{{ toast.message }}</div>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.2s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
