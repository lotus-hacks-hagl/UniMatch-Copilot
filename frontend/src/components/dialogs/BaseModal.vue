<script setup>
import { computed, nextTick, onBeforeUnmount, watch } from 'vue'

const props = defineProps({
  open: { type: Boolean, default: false },
  titleId: { type: String, default: '' },
  descriptionId: { type: String, default: '' },
  closeOnBackdrop: { type: Boolean, default: true },
  closeOnEsc: { type: Boolean, default: true },
  zIndex: { type: Number, default: 110 }
})

const emit = defineEmits(['close'])

const wrapperStyle = computed(() => ({ zIndex: props.zIndex }))
let prevOverflow = ''
let prevActive = null

const focusFirst = async () => {
  await nextTick()
  const el = document.querySelector('[data-modal-root="true"]')
  if (!el) return
  const focusables = el.querySelectorAll(
    'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
  )
  if (focusables.length > 0) focusables[0].focus()
}

const close = () => emit('close')

const onKeyDown = (e) => {
  if (!props.open) return
  if (e.key === 'Escape' && props.closeOnEsc) {
    e.preventDefault()
    close()
  }
}

watch(
  () => props.open,
  async (v) => {
    if (v) {
      prevActive = document.activeElement
      prevOverflow = document.body.style.overflow
      document.body.style.overflow = 'hidden'
      window.addEventListener('keydown', onKeyDown)
      await focusFirst()
    } else {
      document.body.style.overflow = prevOverflow || ''
      window.removeEventListener('keydown', onKeyDown)
      if (prevActive && typeof prevActive.focus === 'function') prevActive.focus()
      prevActive = null
    }
  },
  { immediate: true }
)

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeyDown)
  document.body.style.overflow = prevOverflow || ''
})
</script>

<template>
  <Teleport to="body">
    <Transition name="fade">
      <div
        v-if="open"
        class="fixed inset-0 flex items-center justify-center p-4"
        :style="wrapperStyle"
      >
        <div
          class="absolute inset-0 bg-black/40 backdrop-blur-sm"
          @click="closeOnBackdrop ? close() : null"
        ></div>

        <div
          data-modal-root="true"
          class="relative w-full max-w-lg bg-white rounded-[24px] shadow-2xl border border-black/5"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="titleId || null"
          :aria-describedby="descriptionId || null"
        >
          <slot />
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

