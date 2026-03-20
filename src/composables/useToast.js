import { ref } from 'vue'

const toasts = ref([])
let idCounter = 0

export function useToast() {
  const addToast = (message, type = 'info') => {
    const id = idCounter++
    toasts.value.push({ id, message, type })
    setTimeout(() => {
      toasts.value = toasts.value.filter(t => t.id !== id)
    }, 3000)
  }

  return { toasts, addToast }
}
