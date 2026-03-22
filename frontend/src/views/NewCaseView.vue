<script setup>
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../services/api'

const router = useRouter()
const currentStep = ref(1)
const isSubmitting = ref(false)
const errors = ref({})

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
  target_intake: '',
  background_text: ''
})

const GPA_SCALE_OPTIONS = ['4.0', '10.0', '100.0']
const COUNTRY_OPTIONS = ['USA', 'UK', 'Canada', 'Australia', 'Netherlands', 'Singapore']
const MAJOR_OPTIONS = [
  'Computer Science',
  'Business Administration',
  'Engineering',
  'Arts & Design',
  'Law',
  'Medicine'
]
const INTAKE_OPTIONS = ['Fall 2026', 'Spring 2027', 'Fall 2027']

const gpaScaleNumber = computed(() => parseFloat(form.value.gpa_scale) || 4.0)
const normalizedGpaPreview = computed(() => {
  const raw = parseNumber(form.value.gpa_raw)
  if (raw === null) return null
  const normalized = normalizeGpa(raw, gpaScaleNumber.value)
  if (normalized === null) return null
  return normalized.toFixed(2)
})

function parseNumber(value) {
  if (value === '' || value === null || value === undefined) return null
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : null
}

function normalizeGpa(raw, scale) {
  if (!Number.isFinite(raw) || !Number.isFinite(scale) || scale <= 0) return null
  const normalized = (raw / scale) * 4
  if (!Number.isFinite(normalized)) return null
  return Number(normalized.toFixed(2))
}

function setFieldError(field, message) {
  errors.value = { ...errors.value, [field]: message }
}

function clearFieldError(field) {
  if (!errors.value[field]) return
  const next = { ...errors.value }
  delete next[field]
  errors.value = next
}

function clearAllErrors() {
  errors.value = {}
}

function validateStep(step = currentStep.value) {
  clearAllErrors()

  const raw = parseNumber(form.value.gpa_raw)
  const scale = gpaScaleNumber.value
  const normalized = raw === null ? null : normalizeGpa(raw, scale)
  const ielts = parseNumber(form.value.ielts_overall)
  const sat = parseNumber(form.value.sat_total)
  const budget = parseNumber(String(form.value.budget_usd_per_year).replace(/\D/g, ''))

  if (step >= 1) {
    if (!form.value.full_name.trim()) {
      setFieldError('full_name', 'Full name is required.')
    }
    if (raw !== null) {
      if (raw < 0) {
        setFieldError('gpa_raw', 'GPA cannot be negative.')
      } else if (raw > scale) {
        setFieldError('gpa_raw', `GPA cannot be greater than the selected scale (${scale}).`)
      }
      
      if (normalized === null) {
        setFieldError('gpa_normalized', 'Unable to normalize GPA.')
      } else if (normalized < 0 || normalized > 4) {
        setFieldError('gpa_normalized', 'Normalized GPA must stay between 0.00 and 4.00.')
      }
    }
    
    if (ielts !== null && (ielts < 0 || ielts > 9)) {
      setFieldError('ielts_overall', 'IELTS must be between 0 and 9.')
    }
    if (sat !== null && (sat < 400 || sat > 1600)) {
      setFieldError('sat_total', 'SAT must be between 400 and 1600.')
    }
  }

  if (step >= 2) {
    if (!form.value.intended_major) {
      setFieldError('intended_major', 'Desired major is required.')
    }
    if (form.value.preferred_countries.length === 0) {
      setFieldError('preferred_countries', 'Select at least one preferred country.')
    }
    if (budget === null || budget <= 0) {
      setFieldError('budget_usd_per_year', 'Annual budget must be greater than 0.')
    }
  }

  if (step >= 3) {
    if (!form.value.target_intake) {
      setFieldError('target_intake', 'Target intake is required.')
    }
  }

  return Object.keys(errors.value).length === 0
}

function clampGpaRaw() {
  const raw = parseNumber(form.value.gpa_raw)
  const scale = gpaScaleNumber.value
  if (raw === null) return
  if (raw < 0) form.value.gpa_raw = '0'
  if (raw > scale) form.value.gpa_raw = String(scale)
}

function clampIelts() {
  const value = parseNumber(form.value.ielts_overall)
  if (value === null) return
  if (value < 0) form.value.ielts_overall = '0'
  if (value > 9) form.value.ielts_overall = '9'
}

function clampSat() {
  const value = parseNumber(form.value.sat_total)
  if (value === null) return
  if (value < 400) form.value.sat_total = '400'
  if (value > 1600) form.value.sat_total = '1600'
}

function sanitizeBudget() {
  form.value.budget_usd_per_year = String(form.value.budget_usd_per_year || '').replace(/[^\d]/g, '')
}

const nextStep = () => {
  if (!validateStep(currentStep.value)) {
    return
  }
  if (currentStep.value < 3) currentStep.value++
}

const prevStep = () => {
  if (currentStep.value > 1) currentStep.value--
}

const submitForm = async () => {
  if (!validateStep(3)) {
    const firstErrorStep = errors.value.full_name || errors.value.gpa_raw || errors.value.gpa_normalized || errors.value.ielts_overall || errors.value.sat_total || errors.value.scores
      ? 1
      : (errors.value.intended_major || errors.value.preferred_countries || errors.value.budget_usd_per_year ? 2 : 3)
    currentStep.value = firstErrorStep
    return
  }

  isSubmitting.value = true
  try {
    const raw = parseNumber(form.value.gpa_raw)
    const scale = gpaScaleNumber.value
    const normalized = raw === null ? null : normalizeGpa(raw, scale)

    const payload = {
      ...form.value,
      full_name: form.value.full_name.trim(),
      gpa_raw: raw || 0,
      gpa_scale: scale,
      gpa_normalized: normalized,
      ielts_overall: parseNumber(form.value.ielts_overall),
      sat_total: parseNumber(form.value.sat_total),
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
    const backendMessage = error.response?.data?.error?.details || error.response?.data?.error?.message
    alert(backendMessage || 'Failed to submit form.')
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
            <input v-model.trim="form.full_name" @input="clearFieldError('full_name')" data-testid="new-case-full-name" type="text" class="w-full px-4 py-3 rounded-xl border text-[14px] outline-none focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] bg-[#fafafa] focus:bg-white transition-all placeholder-[#a8a79d]" :class="errors.full_name ? 'border-red-300' : 'border-black/10'" placeholder="e.g. Nguyen Van A" />
            <p v-if="errors.full_name" class="mt-2 text-[12px] text-[#a32d2d]">{{ errors.full_name }}</p>
          </div>
          <div class="grid grid-cols-2 gap-5">
            <div>
              <label class="block text-[13px] font-bold text-[#18180f] mb-2">GPA Raw</label>
              <div class="flex gap-2">
                <input v-model="form.gpa_raw" @input="clearFieldError('gpa_raw'); clearFieldError('gpa_normalized')" @blur="clampGpaRaw()" data-testid="new-case-gpa-raw" type="number" step="0.1" min="0" :max="gpaScaleNumber" class="w-full px-4 py-3 rounded-xl border text-[14px] outline-none focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] bg-[#fafafa] focus:bg-white transition-all placeholder-[#a8a79d]" :class="errors.gpa_raw || errors.gpa_normalized ? 'border-red-300' : 'border-black/10'" placeholder="3.8" />
                <select v-model="form.gpa_scale" @change="clampGpaRaw(); clearFieldError('gpa_raw'); clearFieldError('gpa_normalized')" class="w-28 px-3 py-3 rounded-xl border border-black/10 text-[14px] outline-none focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] bg-[#fafafa] focus:bg-white transition-all font-medium">
                  <option v-for="option in GPA_SCALE_OPTIONS" :key="option" :value="option">/ {{ option }}</option>
                </select>
              </div>
              <p class="mt-2 text-[12px] text-[#6b6a62]">Normalized GPA: <span class="font-bold text-[#18180f]">{{ normalizedGpaPreview ?? 'N/A' }}</span> / 4.00</p>
              <p v-if="errors.gpa_raw" class="mt-1 text-[12px] text-[#a32d2d]">{{ errors.gpa_raw }}</p>
              <p v-else-if="errors.gpa_normalized" class="mt-1 text-[12px] text-[#a32d2d]">{{ errors.gpa_normalized }}</p>
            </div>
            <div>
              <label class="block text-[13px] font-bold text-[#18180f] mb-2">IELTS Score</label>
              <input v-model="form.ielts_overall" @input="clearFieldError('ielts_overall'); clearFieldError('scores')" @blur="clampIelts()" data-testid="new-case-ielts" type="number" step="0.5" min="0" max="9.0" class="w-full px-4 py-3 rounded-xl border text-[14px] outline-none focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] bg-[#fafafa] focus:bg-white transition-all placeholder-[#a8a79d]" :class="errors.ielts_overall || errors.scores ? 'border-red-300' : 'border-black/10'" placeholder="7.5" />
              <p v-if="errors.ielts_overall" class="mt-2 text-[12px] text-[#a32d2d]">{{ errors.ielts_overall }}</p>
            </div>
            <div class="col-span-2 md:col-span-1">
              <label class="block text-[13px] font-bold text-[#18180f] mb-2">SAT Total</label>
              <input v-model="form.sat_total" @input="clearFieldError('sat_total'); clearFieldError('scores')" @blur="clampSat()" data-testid="new-case-sat" type="number" step="10" min="400" max="1600" class="w-full px-4 py-3 rounded-xl border text-[14px] outline-none focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] bg-[#fafafa] focus:bg-white transition-all placeholder-[#a8a79d]" :class="errors.sat_total || errors.scores ? 'border-red-300' : 'border-black/10'" placeholder="1450" />
              <p v-if="errors.sat_total" class="mt-2 text-[12px] text-[#a32d2d]">{{ errors.sat_total }}</p>
            </div>
            <p v-if="errors.scores" class="col-span-2 text-[12px] text-[#a32d2d]">{{ errors.scores }}</p>
          </div>
        </div>

        <!-- Step 2: Profile -->
        <div v-else-if="currentStep === 2" key="step2" class="space-y-6">
          <div>
            <label class="block text-[13px] font-bold text-[#18180f] mb-2">Desired Major</label>
            <select v-model="form.intended_major" @change="clearFieldError('intended_major')" data-testid="new-case-major" class="w-full px-4 py-3 rounded-xl border text-[14px] outline-none focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] bg-[#fafafa] focus:bg-white transition-all font-medium" :class="errors.intended_major ? 'border-red-300' : 'border-black/10'">
              <option disabled value="">Select a major...</option>
              <option v-for="major in MAJOR_OPTIONS" :key="major" :value="major">{{ major }}</option>
            </select>
            <p v-if="errors.intended_major" class="mt-2 text-[12px] text-[#a32d2d]">{{ errors.intended_major }}</p>
          </div>
          <div>
            <label class="block text-[13px] font-bold text-[#18180f] mb-2">Preferred Countries</label>
            <div class="flex flex-wrap gap-2.5">
              <button 
                v-for="c in COUNTRY_OPTIONS" 
                :key="c"
                @click="toggleCountry(c)"
                type="button"
                class="px-4 py-2 rounded-xl text-[13px] font-bold transition-all border-2"
                :class="form.preferred_countries.includes(c) ? 'border-[#a32d2d] bg-red-50 text-[#a32d2d]' : 'border-transparent bg-[#f4f5f7] text-[#6b6a62] hover:bg-gray-200'"
              >
                {{ c }}
              </button>
            </div>
            <p v-if="errors.preferred_countries" class="mt-2 text-[12px] text-[#a32d2d]">{{ errors.preferred_countries }}</p>
          </div>
          <div>
            <label class="block text-[13px] font-bold text-[#18180f] mb-2">Annual Budget (USD)</label>
            <div class="relative">
              <input v-model="form.budget_usd_per_year" @input="sanitizeBudget(); clearFieldError('budget_usd_per_year')" data-testid="new-case-budget" type="text" inputmode="numeric" class="w-full pl-9 pr-4 py-3 rounded-xl border text-[14px] outline-none focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] bg-[#fafafa] focus:bg-white transition-all placeholder-[#a8a79d]" :class="errors.budget_usd_per_year ? 'border-red-300' : 'border-black/10'" placeholder="e.g. 40000" />
              <span class="absolute left-4 top-3 text-[#a8a79d] font-bold text-[14px]">$</span>
            </div>
            <p v-if="errors.budget_usd_per_year" class="mt-2 text-[12px] text-[#a32d2d]">{{ errors.budget_usd_per_year }}</p>
          </div>
          <div>
            <label class="block text-[13px] font-bold text-[#18180f] mb-2">Student Background (Optional)</label>
            <textarea v-model="form.background_text" class="w-full px-4 py-3 rounded-xl border text-[14px] outline-none focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] bg-[#fafafa] focus:bg-white transition-all placeholder-[#a8a79d] min-h-[100px] resize-y" placeholder="e.g. Previous study history, specific interests, or special needs..."></textarea>
          </div>
        </div>

        <!-- Step 3: Target -->
        <div v-else-if="currentStep === 3" key="step3" class="space-y-6">
          <div>
            <label class="block text-[13px] font-bold text-[#18180f] mb-2">Target Intake Term</label>
            <select v-model="form.target_intake" @change="clearFieldError('target_intake')" data-testid="new-case-intake" class="w-full px-4 py-3 rounded-xl border text-[14px] outline-none focus:ring-2 focus:ring-[#a32d2d]/10 focus:border-[#a32d2d] bg-[#fafafa] focus:bg-white transition-all font-medium" :class="errors.target_intake ? 'border-red-300' : 'border-black/10'">
              <option disabled value="">Select intake...</option>
              <option v-for="intake in INTAKE_OPTIONS" :key="intake" :value="intake">{{ intake }}</option>
            </select>
            <p v-if="errors.target_intake" class="mt-2 text-[12px] text-[#a32d2d]">{{ errors.target_intake }}</p>
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
          data-testid="new-case-continue"
          class="btn-primary shadow-[0_4px_14px_rgba(163,45,45,0.35)] shrink-0 w-28 flex justify-center"
        >
          Continue
        </button>
        
        <button 
          v-if="currentStep === 3" 
          @click="submitForm"
          :disabled="isSubmitting"
          data-testid="new-case-submit"
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
