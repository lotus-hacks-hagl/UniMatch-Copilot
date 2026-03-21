<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../services/api'

const router = useRouter()
const currentStep = ref(1)
const isSubmitting = ref(false)

const form = ref({
  full_name: '',
  email: '',
  phone: '',
  gpa_raw: '',
  gpa_scale: '4.0',
  ielts_overall: '',
  intended_major: '',
  preferred_countries: [],
  budget_usd_per_year: '',
  target_intake: ''
})

const nextStep = () => {
  if (currentStep.value < 3) currentStep.value++
}

const prevStep = () => {
  if (currentStep.value > 1) currentStep.value--
}

const submitForm = async () => {
  isSubmitting.value = true
  try {
    // Basic normalization of strings to types
    const payload = {
      ...form.value,
      gpa_raw: parseFloat(form.value.gpa_raw) || 0,
      gpa_scale: parseFloat(form.value.gpa_scale) || 4.0,
      ielts_overall: parseFloat(form.value.ielts_overall) || 0,
      budget_usd_per_year: parseInt(form.value.budget_usd_per_year.replace(/\D/g,'')) || 0,
    }
    const response = await api.post('/cases', payload)
    
    // Redirect to new case details page
    if (response.data && response.data.case_id) {
      router.push('/cases/' + response.data.case_id)
    } else {
      router.push('/cases')
    }
  } catch (error) {
    console.error('Failed to create case:', error)
    alert('Failed to submit form.')
  } finally {
    isSubmitting.value = false
  }
}

const toggleCountry = (code) => {
  const idx = form.value.preferred_countries.indexOf(code)
  if (idx > -1) {
    form.value.preferred_countries.splice(idx, 1)
  } else {
    form.value.preferred_countries.push(code)
  }
}
</script>

<template>
  <div class="max-w-2xl mx-auto px-5 py-8">
    <div class="bg-surface rounded-xl border border-black/5 shadow-sm p-8">
      <!-- Stepper -->
      <div class="flex items-center justify-between mb-8 relative">
        <div class="absolute left-0 top-1/2 -translate-y-1/2 w-full h-0.5 bg-gray-200 z-0"></div>
        <div 
          class="absolute left-0 top-1/2 -translate-y-1/2 h-0.5 bg-primary transition-all duration-300 z-0"
          :style="{ width: ((currentStep - 1) * 50) + '%' }"
        ></div>
        
        <div 
          v-for="step in [1, 2, 3]" 
          :key="step"
          class="w-10 h-10 rounded-full flex items-center justify-center font-medium text-[13px] relative z-10 transition-colors duration-300"
          :class="step <= currentStep ? 'bg-primary text-white border-2 border-surface' : 'bg-gray-100 text-gray-400 border-2 border-surface'"
        >
          {{ step }}
        </div>
      </div>

      <div class="mb-8 text-center">
        <h2 class="text-xl font-bold text-text mb-1">
          {{ currentStep === 1 ? 'Student Information' : (currentStep === 2 ? 'Academic Profile' : 'Target Preferences') }}
        </h2>
        <p class="text-[13px] text-text-muted">
          {{ currentStep === 1 ? 'Enter basic contact and academic scores.' : (currentStep === 2 ? 'Define the desired major and budget.' : 'Set the target intake and finalize.') }}
        </p>
      </div>

      <!-- Step 1: Student -->
      <div v-if="currentStep === 1" class="space-y-4 animate-fade-in">
        <div>
          <label class="block text-[13px] font-medium text-text mb-1.5">Full Name</label>
          <input v-model="form.full_name" type="text" class="w-full px-3.5 py-2.5 rounded-lg border-black/10 text-[13px] focus:ring-1 focus:ring-primary focus:border-primary bg-bg text-text transition-shadow" placeholder="e.g. Nguyen Van A" />
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-[13px] font-medium text-text mb-1.5">GPA Raw</label>
            <div class="flex gap-2">
              <input v-model="form.gpa_raw" type="number" step="0.1" class="w-full px-3.5 py-2.5 rounded-lg border-black/10 text-[13px] focus:ring-1 focus:ring-primary focus:border-primary bg-bg text-text transition-shadow" placeholder="3.8" />
              <select v-model="form.gpa_scale" class="w-24 px-2 py-2.5 rounded-lg border-black/10 text-[13px] focus:ring-1 focus:ring-primary focus:border-primary bg-bg text-text">
                <option value="4.0">/ 4.0</option>
                <option value="10.0">/ 10</option>
                <option value="100.0">/ 100</option>
              </select>
            </div>
          </div>
          <div>
            <label class="block text-[13px] font-medium text-text mb-1.5">IELTS Score</label>
            <input v-model="form.ielts_overall" type="number" step="0.5" max="9.0" class="w-full px-3.5 py-2.5 rounded-lg border-black/10 text-[13px] focus:ring-1 focus:ring-primary focus:border-primary bg-bg text-text transition-shadow" placeholder="7.5" />
          </div>
        </div>
      </div>

      <!-- Step 2: Profile -->
      <div v-if="currentStep === 2" class="space-y-4 animate-fade-in">
        <div>
          <label class="block text-[13px] font-medium text-text mb-1.5">Desired Major</label>
          <select v-model="form.intended_major" class="w-full px-3.5 py-2.5 rounded-lg border-black/10 text-[13px] focus:ring-1 focus:ring-primary focus:border-primary bg-bg text-text transition-shadow">
            <option disabled value="">Select a major...</option>
            <option value="Computer Science">Computer Science</option>
            <option value="Business Administration">Business Administration</option>
            <option value="Engineering">Engineering</option>
            <option value="Arts & Design">Arts & Design</option>
            <option value="Law">Law</option>
            <option value="Medicine">Medicine</option>
          </select>
        </div>
        <div>
          <label class="block text-[13px] font-medium text-text mb-1.5">Preferred Countries</label>
          <div class="flex flex-wrap gap-2">
            <button 
              v-for="c in ['USA', 'UK', 'Canada', 'Australia', 'Netherlands', 'Singapore']" 
              :key="c"
              @click="toggleCountry(c)"
              class="px-3 py-1.5 rounded-lg border text-[13px] transition-colors"
              :class="form.preferred_countries.includes(c) ? 'border-primary bg-primary/10 text-primary font-medium' : 'border-black/10 text-text hover:bg-bg'"
            >
              {{ c }}
            </button>
          </div>
        </div>
        <div>
          <label class="block text-[13px] font-medium text-text mb-1.5">Annual Budget (USD)</label>
          <div class="relative">
            <input v-model="form.budget_usd_per_year" type="text" class="w-full pl-8 pr-3.5 py-2.5 rounded-lg border-black/10 text-[13px] focus:ring-1 focus:ring-primary focus:border-primary bg-bg text-text transition-shadow" placeholder="e.g. 40,000" />
            <span class="absolute left-3.5 top-2.5 text-text-muted text-[13px]">$</span>
          </div>
        </div>
      </div>

      <!-- Step 3: Target -->
      <div v-if="currentStep === 3" class="space-y-4 animate-fade-in">
        <div>
          <label class="block text-[13px] font-medium text-text mb-1.5">Target Intake Term</label>
          <select v-model="form.target_intake" class="w-full px-3.5 py-2.5 rounded-lg border-black/10 text-[13px] focus:ring-1 focus:ring-primary focus:border-primary bg-bg text-text transition-shadow">
            <option disabled value="">Select intake...</option>
            <option value="Fall 2026">Fall 2026</option>
            <option value="Spring 2027">Spring 2027</option>
            <option value="Fall 2027">Fall 2027</option>
          </select>
        </div>
        <div class="p-4 bg-secondary rounded-lg border border-primary/20">
          <div class="flex items-start gap-3">
            <div class="w-8 h-8 rounded-full bg-primary text-white flex items-center justify-center shrink-0">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
            </div>
            <div>
              <h4 class="text-[13px] font-medium text-primary">Ready to process</h4>
              <p class="text-[12px] text-primary/80 mt-1">Once submitted, TinyFish AI will immediately analyze this profile against the University KB to generate Safe, Match, and Reach recommendations.</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Actions -->
      <div class="mt-8 flex items-center justify-between pt-5 border-t border-black/5">
        <button 
          v-if="currentStep > 1" 
          @click="prevStep"
          class="px-5 py-2.5 rounded-lg text-[13px] font-medium border border-black/10 text-text hover:bg-bg transition-colors"
        >
          Back
        </button>
        <div v-else></div> <!-- Spacer -->
        
        <button 
          v-if="currentStep < 3" 
          @click="nextStep"
          class="px-5 py-2.5 rounded-lg text-[13px] font-medium bg-primary text-white hover:bg-primary-hover transition-colors shadow-sm"
        >
          Continue
        </button>
        
        <button 
          v-if="currentStep === 3" 
          @click="submitForm"
          :disabled="isSubmitting"
          class="px-5 py-2.5 rounded-lg text-[13px] font-medium bg-primary text-white hover:bg-primary-hover transition-colors shadow-sm disabled:opacity-50"
        >
          {{ isSubmitting ? 'Creating Case...' : 'Submit & Analyze' }}
        </button>
      </div>

    </div>
  </div>
</template>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.3s ease-out forwards;
}
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(5px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
