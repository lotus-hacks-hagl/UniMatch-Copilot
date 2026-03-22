import { reactive } from 'vue'

const state = reactive({
  open: false,
  title: '',
  message: '',
  variant: 'default',
  confirmText: 'Confirm',
  cancelText: 'Cancel',
  loading: false,
  _resolve: null
})

function closeWith(value) {
  const r = state._resolve
  state.open = false
  state.loading = false
  state._resolve = null
  if (typeof r === 'function') r(value)
}

export function useDialog() {
  const confirm = ({ title, message, variant = 'default', confirmText = 'Confirm', cancelText = 'Cancel' }) => {
    state.title = title || ''
    state.message = message || ''
    state.variant = variant
    state.confirmText = confirmText
    state.cancelText = cancelText
    state.loading = false
    state.open = true
    return new Promise((resolve) => {
      state._resolve = resolve
    })
  }

  const onConfirm = () => closeWith(true)
  const onCancel = () => closeWith(false)
  const onClose = () => closeWith(false)

  return { state, confirm, onConfirm, onCancel, onClose }
}

