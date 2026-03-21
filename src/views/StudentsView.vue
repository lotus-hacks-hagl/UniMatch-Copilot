<script setup>
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
</script>
<template>
  <div class="px-5 py-5 space-y-4">
    <h2 class="text-lg font-medium">{{ t('students.title') }}</h2>
    <p class="text-text-muted">{{ t('students.placeholder') }}</p>
  </div>
</template>
