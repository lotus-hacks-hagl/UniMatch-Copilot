<template>
  <AppShell v-if="!route.meta.public">
    <RouterView />
  </AppShell>
  <RouterView v-else />
  <ToastStack />
</template>

<script setup>
import { RouterView, useRoute } from 'vue-router'
import AppShell from './components/AppShell.vue'
import ToastStack from './components/ToastStack.vue'

const route = useRoute()
</script>
