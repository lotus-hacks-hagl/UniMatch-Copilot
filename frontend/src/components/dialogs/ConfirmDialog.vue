<script setup>
import { computed } from 'vue'
import BaseModal from './BaseModal.vue'

const props = defineProps({
  open: { type: Boolean, default: false },
  title: { type: String, default: '' },
  message: { type: String, default: '' },
  variant: { type: String, default: 'default' },
  confirmText: { type: String, default: 'Confirm' },
  cancelText: { type: String, default: 'Cancel' },
  loading: { type: Boolean, default: false }
})

const emit = defineEmits(['confirm', 'cancel', 'close'])

const titleId = computed(() => 'dlg_title_' + (props.variant || 'default'))
const descId = computed(() => 'dlg_desc_' + (props.variant || 'default'))
const confirmClass = computed(() => (props.variant === 'danger' ? 'btn-danger' : 'btn-primary'))
</script>

<template>
  <BaseModal :open="open" :title-id="titleId" :description-id="descId" @close="$emit('close')">
    <div class="p-8">
      <div class="flex items-start justify-between gap-4">
        <div>
          <div :id="titleId" class="text-[18px] font-extrabold text-[#18180f]">
            {{ title }}
          </div>
          <div :id="descId" class="mt-2 text-[14px] text-[#6b6a62] leading-relaxed whitespace-pre-wrap">
            {{ message }}
          </div>
        </div>
        <button class="btn-outline px-3 py-2" @click="$emit('close')">✕</button>
      </div>

      <div class="mt-7 flex items-center justify-end gap-2">
        <button class="btn-outline" :disabled="loading" @click="$emit('cancel')">
          {{ cancelText }}
        </button>
        <button :class="confirmClass" :disabled="loading" @click="$emit('confirm')">
          {{ loading ? '...' : confirmText }}
        </button>
      </div>
    </div>
  </BaseModal>
</template>

