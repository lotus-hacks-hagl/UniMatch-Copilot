<template>
  <AppShell v-if="!route.meta.public">
    <RouterView />
  </AppShell>
  <RouterView v-else />
</template>

<script setup>
import { RouterView, useRoute } from 'vue-router'
import AppShell from './components/AppShell.vue'

const route = useRoute()
</script>
