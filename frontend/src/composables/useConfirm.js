import { ref, reactive } from 'vue'

const state = reactive({
  show: false,
  title: '',
  message: '',
  confirmLabel: 'Confirm',
  cancelLabel: 'Cancel',
  type: 'primary', // 'primary', 'danger', 'info'
  resolve: null,
  reject: null
})

export function useConfirm() {
  const confirm = (options = {}) => {
    state.title = options.title || 'Are you sure?'
    state.message = options.message || ''
    state.confirmLabel = options.confirmLabel || 'Confirm'
    state.cancelLabel = options.cancelLabel || 'Cancel'
    state.type = options.type || 'primary'
    state.show = true

    return new Promise((resolve) => {
      state.resolve = resolve
    })
  }

  const handleConfirm = () => {
    state.show = false
    if (state.resolve) state.resolve(true)
  }

  const handleCancel = () => {
    state.show = false
    if (state.resolve) state.resolve(false)
  }

  return {
    state,
    confirm,
    handleConfirm,
    handleCancel
  }
}
