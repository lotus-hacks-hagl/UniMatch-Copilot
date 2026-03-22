<script setup>
import ConfirmDialog from './ConfirmDialog.vue'
import { useDialog } from '../../composables/useDialog'

const dlg = useDialog()
</script>

<template>
  <ConfirmDialog
    :open="dlg.state.open"
    :title="dlg.state.title"
    :message="dlg.state.message"
    :variant="dlg.state.variant"
    :confirm-text="dlg.state.confirmText"
    :cancel-text="dlg.state.cancelText"
    :loading="dlg.state.loading"
    @confirm="dlg.onConfirm"
    @cancel="dlg.onCancel"
    @close="dlg.onClose"
  />
</template>

