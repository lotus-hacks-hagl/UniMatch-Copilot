<script setup>
</script>
<template>
  <div class="px-7 py-6 max-w-[1200px] mx-auto space-y-6">
    <h2 class="text-xl font-bold text-text">Counselor Analytics</h2>

    <div class="grid grid-cols-4 gap-4">
      <div class="bg-surface rounded-xl p-5 border border-black/5 shadow-sm">
        <div class="text-[13px] text-text-muted mb-1">Placement Rate</div>
        <div class="text-2xl font-bold text-text">92.4%</div>
      </div>
      <div class="bg-surface rounded-xl p-5 border border-black/5 shadow-sm">
        <div class="text-[13px] text-text-muted mb-1">Avg Scholarship</div>
        <div class="text-2xl font-bold text-text">$14,500</div>
      </div>
      <div class="bg-surface rounded-xl p-5 border border-black/5 shadow-sm">
        <div class="text-[13px] text-text-muted mb-1">Time Saved by AI</div>
        <div class="text-2xl font-bold text-text">142 hrs</div>
      </div>
      <div class="bg-surface rounded-xl p-5 border border-black/5 shadow-sm bg-secondary border-primary/20">
        <div class="text-[13px] text-primary mb-1">Client Satisfaction</div>
        <div class="text-2xl font-bold text-primary">4.9/5</div>
      </div>
    </div>

    <div class="grid grid-cols-2 gap-6 mt-6">
      <div class="bg-surface rounded-xl border border-black/5 shadow-sm p-6">
        <h3 class="text-sm font-bold text-text mb-6">Placements by Region</h3>
        <div class="space-y-4">
          <div>
            <div class="flex justify-between text-[12px] text-text-muted mb-1"><span>United States</span><span>64%</span></div>
            <div class="w-full h-2 bg-gray-100 rounded-full"><div class="h-full bg-primary rounded-full" style="width: 64%"></div></div>
          </div>
          <div>
            <div class="flex justify-between text-[12px] text-text-muted mb-1"><span>United Kingdom</span><span>21%</span></div>
            <div class="w-full h-2 bg-gray-100 rounded-full"><div class="h-full bg-primary/80 rounded-full" style="width: 21%"></div></div>
          </div>
          <div>
            <div class="flex justify-between text-[12px] text-text-muted mb-1"><span>Canada</span><span>10%</span></div>
            <div class="w-full h-2 bg-gray-100 rounded-full"><div class="h-full bg-primary/60 rounded-full" style="width: 10%"></div></div>
          </div>
          <div>
            <div class="flex justify-between text-[12px] text-text-muted mb-1"><span>Australia</span><span>5%</span></div>
            <div class="w-full h-2 bg-gray-100 rounded-full"><div class="h-full bg-primary/40 rounded-full" style="width: 5%"></div></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
