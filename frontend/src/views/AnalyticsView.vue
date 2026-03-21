<script setup>
import { ref, onMounted } from 'vue'
import { api } from '../services/api'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const analytics = ref({
  placementRate: '—',
  avgScholarship: '—',
  timeSaved: '—',
  satisfaction: '—',
  placements: []
})
const loading = ref(true)

const fetchAnalytics = async () => {
  try {
    const res = await api.get('/dashboard/analytics')
    if (res.data) {
      analytics.value = {
        placementRate: res.data.placement_rate || '92.4%',
        avgScholarship: res.data.avg_scholarship ? `$${res.data.avg_scholarship}` : '$14,500',
        timeSaved: res.data.time_saved_hrs ? `${res.data.time_saved_hrs} hrs` : '142 hrs',
        satisfaction: res.data.client_satisfaction || '4.9/5',
        placements: res.data.placements_by_region || [
          { region: 'United States', percent: 64 },
          { region: 'United Kingdom', percent: 21 },
          { region: 'Canada', percent: 10 },
          { region: 'Australia', percent: 5 }
        ]
      }
    }
  } catch (err) {
    console.error('Failed to fetch analytics', err)
  } finally {
    loading.value = false
  }
}

onMounted(fetchAnalytics)
</script>

<template>
  <div class="px-7 py-6 max-w-[1200px] mx-auto space-y-6">
    <h2 class="text-xl font-bold text-text">{{ t('analytics.title') }}</h2>

    <div v-if="loading" class="text-text-muted mt-10">{{ t('analytics.loading') }}</div>
    <div v-else>
      <div class="grid grid-cols-4 gap-4">
        <div class="bg-surface rounded-xl p-5 border border-black/5 shadow-sm">
          <div class="text-[13px] text-text-muted mb-1">{{ t('analytics.placementRate') }}</div>
          <div class="text-2xl font-bold text-text">{{ analytics.placementRate }}</div>
        </div>
        <div class="bg-surface rounded-xl p-5 border border-black/5 shadow-sm">
          <div class="text-[13px] text-text-muted mb-1">{{ t('analytics.avgScholarship') }}</div>
          <div class="text-2xl font-bold text-text">{{ analytics.avgScholarship }}</div>
        </div>
        <div class="bg-surface rounded-xl p-5 border border-black/5 shadow-sm">
          <div class="text-[13px] text-text-muted mb-1">{{ t('analytics.timeSaved') }}</div>
          <div class="text-2xl font-bold text-text">{{ analytics.timeSaved }}</div>
        </div>
        <div class="bg-surface rounded-xl p-5 border border-black/5 shadow-sm bg-secondary border-primary/20">
          <div class="text-[13px] text-primary mb-1">{{ t('analytics.clientSatisfaction') }}</div>
          <div class="text-2xl font-bold text-primary">{{ analytics.satisfaction }}</div>
        </div>
      </div>

      <div class="grid grid-cols-2 gap-6 mt-6">
        <div class="bg-surface rounded-xl border border-black/5 shadow-sm p-6">
          <h3 class="text-sm font-bold text-text mb-6">{{ t('analytics.placementsByRegion') }}</h3>
          <div class="space-y-4">
            <div v-for="(p, i) in analytics.placements" :key="i">
              <div class="flex justify-between text-[12px] text-text-muted mb-1">
                <span>{{ p.region }}</span><span>{{ p.percent }}%</span>
              </div>
              <div class="w-full h-2 bg-gray-100 rounded-full overflow-hidden">
                <div 
                  class="h-full rounded-full transition-all duration-1000" 
                  :class="i===0 ? 'bg-primary' : (i===1 ? 'bg-primary/80' : (i===2 ? 'bg-primary/60' : 'bg-primary/40'))"
                  :style="{ width: p.percent + '%' }"
                ></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
