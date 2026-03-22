<template>
  <AppShell v-if="!route.meta.public">
    <RouterView />
  </AppShell>
  <RouterView v-else />
  <ToastStack />
  <DialogHost />
</template>

<script setup>
import { RouterView, useRoute } from 'vue-router'
import AppShell from './components/AppShell.vue'
import ToastStack from './components/ToastStack.vue'
import DialogHost from './components/dialogs/DialogHost.vue'

const route = useRoute()
</script>
