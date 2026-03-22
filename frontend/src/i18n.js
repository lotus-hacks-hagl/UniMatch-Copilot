import { createI18n } from 'vue-i18n'
import en from './locales/en.json'
import vi from './locales/vi.json'

const i18n = createI18n({
  legacy: false, // Vue 3 Composition API
  globalInjection: true,
  locale: 'en',
  fallbackLocale: 'vi',
  messages: {
    en,
    vi
  }
})

export default i18n
