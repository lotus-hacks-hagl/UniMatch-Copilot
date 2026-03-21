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
  sat_total: '',
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
  if (!form.value.ielts_overall && !form.value.sat_total) {
    alert('Please provide either IELTS or SAT score.')
    currentStep.value = 1
    return
  }

  isSubmitting.value = true
  try {
    const raw = parseFloat(form.value.gpa_raw) || 0
    const scale = parseFloat(form.value.gpa_scale) || 4.0
    const normalized = (raw / scale) * 4.0

    const payload = {
      ...form.value,
      gpa_raw: raw,
      gpa_scale: scale,
      gpa_normalized: parseFloat(normalized.toFixed(2)),
      ielts_overall: parseFloat(form.value.ielts_overall) || null,
      sat_total: parseInt(form.value.sat_total) || null,
      budget_usd_per_year: parseInt(form.value.budget_usd_per_year.toString().replace(/\D/g,'')) || 0,
    }
    const response = await api.post('/cases', payload)
    
    if (response.data?.success && response.data?.data?.case_id) {
      router.push('/cases/' + response.data.data.case_id)
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
  <div class="px-8 py-8 w-full max-w-3xl mx-auto font-sans">
    <div class="card-soft p-10 md:p-12 relative overflow-hidden">
      <!-- Decorative background glow -->
      <div class="absolute -top-32 -right-32 w-64 h-64 bg-red-500/5 rounded-full blur-3xl pointer-events-none"></div>

      <!-- Stepper -->
      <div class="flex items-center justify-between mb-10 relative">
        <div class="absolute left-0 top-1/2 -translate-y-1/2 w-full h-1 bg-gray-100 rounded-full z-0"></div>
        <div 
          class="absolute left-0 top-1/2 -translate-y-1/2 h-1 bg-[#a32d2d] rounded-full transition-all duration-700 ease-out z-0"
          :style="{ width: ((currentStep - 1) * 50) + '%' }"
        ></div>
        
        <div 
          v-for="step in [1, 2, 3]" 
          :key="step"
          class="w-12 h-12 rounded-full flex items-center justify-center font-bold text-[15px] relative z-10 transition-all duration-500 shadow-sm"
          :class="step <= currentStep ? 'bg-[#a32d2d] text-white border-4 border-white shadow-[0_4px_10px_rgba(163,45,45,0.2)]' : 'bg-white text-[#a8a79d] border-4 border-gray-50'"
        >
          {{ step }}
        </div>
      </div>

      <div class="mb-10 text-center animate-fade-in-up">
        <h2 class="text-2xl font-bold text-[#18180f] mb-2">
          {{ currentStep === 1 ? 'Student Information' : (currentStep === 2 ? 'Academic Profile' : 'Target Preferences') }}
        </h2>
        <p class="text-[14px] text-[#6b6a62]">
          {{ currentStep === 1 ? 'Enter basic contact and academic scores to begin.' : (currentStep === 2 ? 'Define the desired major and budget constraints.' : 'Set the target intake and finalize for AI analysis.') }}
        </p>
      </div>

      <Transition name="fade" mode="out-in">
        <!-- Step 1: Student -->
        <div v-if="currentStep === 1" key="step1" class="space-y-6">
          <div>
            <label class="block text-[13px] font-bold text-[#18180f] mb-2">Full Name</label>
            <input v-model="form.full_name" type="text" class="w-full px-4 py-3 rounded-xl border border-black/10 text-[14px] outline-none focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] bg-[#fafafa] focus:bg-white transition-all placeholder-[#a8a79d]" placeholder="e.g. Nguyen Van A" />
          </div>
          <div class="grid grid-cols-2 gap-5">
            <div>
              <label class="block text-[13px] font-bold text-[#18180f] mb-2">GPA Raw</label>
              <div class="flex gap-2">
                <input v-model="form.gpa_raw" type="number" step="0.1" class="w-full px-4 py-3 rounded-xl border border-black/10 text-[14px] outline-none focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] bg-[#fafafa] focus:bg-white transition-all placeholder-[#a8a79d]" placeholder="3.8" />
                <select v-model="form.gpa_scale" class="w-28 px-3 py-3 rounded-xl border border-black/10 text-[14px] outline-none focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] bg-[#fafafa] focus:bg-white transition-all font-medium">
                  <option value="4.0">/ 4.0</option>
                  <option value="10.0">/ 10</option>
                  <option value="100.0">/ 100</option>
                </select>
              </div>
            </div>
            <div>
              <label class="block text-[13px] font-bold text-[#18180f] mb-2">IELTS Score</label>
              <input v-model="form.ielts_overall" type="number" step="0.5" max="9.0" class="w-full px-4 py-3 rounded-xl border border-black/10 text-[14px] outline-none focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] bg-[#fafafa] focus:bg-white transition-all placeholder-[#a8a79d]" placeholder="7.5" />
            </div>
            <div class="col-span-2 md:col-span-1">
              <label class="block text-[13px] font-bold text-[#18180f] mb-2">SAT Total</label>
              <input v-model="form.sat_total" type="number" step="10" min="400" max="1600" class="w-full px-4 py-3 rounded-xl border border-black/10 text-[14px] outline-none focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] bg-[#fafafa] focus:bg-white transition-all placeholder-[#a8a79d]" placeholder="1450" />
            </div>
          </div>
        </div>

        <!-- Step 2: Profile -->
        <div v-else-if="currentStep === 2" key="step2" class="space-y-6">
          <div>
            <label class="block text-[13px] font-bold text-[#18180f] mb-2">Desired Major</label>
            <select v-model="form.intended_major" class="w-full px-4 py-3 rounded-xl border border-black/10 text-[14px] outline-none focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] bg-[#fafafa] focus:bg-white transition-all font-medium">
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
            <label class="block text-[13px] font-bold text-[#18180f] mb-2">Preferred Countries</label>
            <div class="flex flex-wrap gap-2.5">
              <button 
                v-for="c in ['USA', 'UK', 'Canada', 'Australia', 'Netherlands', 'Singapore']" 
                :key="c"
                @click="toggleCountry(c)"
                class="px-4 py-2 rounded-xl text-[13px] font-bold transition-all border-2"
                :class="form.preferred_countries.includes(c) ? 'border-[#a32d2d] bg-red-50 text-[#a32d2d]' : 'border-transparent bg-[#f4f5f7] text-[#6b6a62] hover:bg-gray-200'"
              >
                {{ c }}
              </button>
            </div>
          </div>
          <div>
            <label class="block text-[13px] font-bold text-[#18180f] mb-2">Annual Budget (USD)</label>
            <div class="relative">
              <input v-model="form.budget_usd_per_year" type="text" class="w-full pl-9 pr-4 py-3 rounded-xl border border-black/10 text-[14px] outline-none focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] bg-[#fafafa] focus:bg-white transition-all placeholder-[#a8a79d]" placeholder="e.g. 40,000" />
              <span class="absolute left-4 top-3 text-[#a8a79d] font-bold text-[14px]">$</span>
            </div>
          </div>
        </div>

        <!-- Step 3: Target -->
        <div v-else-if="currentStep === 3" key="step3" class="space-y-6">
          <div>
            <label class="block text-[13px] font-bold text-[#18180f] mb-2">Target Intake Term</label>
            <select v-model="form.target_intake" class="w-full px-4 py-3 rounded-xl border border-black/10 text-[14px] outline-none focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] bg-[#fafafa] focus:bg-white transition-all font-medium">
              <option disabled value="">Select intake...</option>
              <option value="Fall 2026">Fall 2026</option>
              <option value="Spring 2027">Spring 2027</option>
              <option value="Fall 2027">Fall 2027</option>
            </select>
          </div>
          
          <!-- Callout -->
          <div class="p-5 mt-4 bg-red-50 rounded-xl border border-red-100 flex items-start gap-4">
            <div class="w-10 h-10 rounded-full bg-white shadow-sm text-[#a32d2d] flex items-center justify-center shrink-0">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path></svg>
            </div>
            <div>
              <h4 class="text-[14px] font-bold text-[#a32d2d]">Ready to process</h4>
              <p class="text-[13px] text-[#a32d2d]/80 mt-1 leading-relaxed">Once submitted, TinyFish AI will immediately analyze this profile against the University KB to generate Safe, Match, and Reach recommendations.</p>
            </div>
          </div>
        </div>
      </Transition>

      <!-- Actions -->
      <div class="mt-12 flex items-center justify-between pt-6 border-t border-black/5">
        <button 
          v-if="currentStep > 1" 
          @click="prevStep"
          class="btn-outline shrink-0 w-28 flex justify-center"
        >
          Back
        </button>
        <div v-else></div> <!-- Spacer -->
        
        <button 
          v-if="currentStep < 3" 
          @click="nextStep"
          class="btn-primary shadow-[0_4px_14px_rgba(163,45,45,0.35)] shrink-0 w-28 flex justify-center"
        >
          Continue
        </button>
        
        <button 
          v-if="currentStep === 3" 
          @click="submitForm"
          :disabled="isSubmitting"
          class="btn-primary shadow-[0_4px_14px_rgba(163,45,45,0.35)] shrink-0 px-8 disabled:opacity-50"
        >
          {{ isSubmitting ? 'Creating Case...' : 'Submit & Analyze' }}
        </button>
      </div>

    </div>
  </div>
</template>

<style scoped>
.animate-fade-in-up {
  animation: fadeInUp 0.4s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
