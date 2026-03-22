<script setup>
defineProps({
  title: { type: String, default: 'Coming soon' },
  description: { type: String, default: 'This feature is under construction.' }
})
</script>

<template>
  <div class="bg-white p-8 rounded-[20px] shadow-sm border border-black/5">
    <div class="flex flex-col items-center justify-center text-center py-10">
      <div class="w-20 h-20 bg-gray-50 rounded-full flex items-center justify-center mb-5">
        <svg class="w-10 h-10 text-[#a8a79d]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"></path>
        </svg>
      </div>
      <div class="text-[16px] font-bold text-[#18180f] mb-1">{{ title }}</div>
      <p class="text-[14px] text-[#6b6a62] max-w-lg">{{ description }}</p>
    </div>
  </div>
</template>

