<template>
  <AppShell v-if="!route.meta.public">
    <RouterView />
  </AppShell>
  <RouterView v-else />
  <ToastStack />
  <ConfirmDialog />
</template>

<script setup>
import { RouterView, useRoute } from 'vue-router'
import AppShell from './components/AppShell.vue'
import ToastStack from './components/ToastStack.vue'
import ConfirmDialog from './components/ConfirmDialog.vue'

const route = useRoute()
</script>
