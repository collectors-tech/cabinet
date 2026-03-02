import i18next from 'i18next'
import LanguageDetector from 'i18next-browser-languagedetector'
import { initReactI18next } from 'react-i18next'
import commonAr from '@/locales/ar/common.json'
import commonEn from '@/locales/en/common.json'
import navEn from '@/locales/en/nav.json'
import pagesEn from '@/locales/en/pages.json'

function applyDocumentLocale(language: string) {
  if (typeof document === 'undefined') {
    return
  }
  const direction = i18next.dir(language)
  document.documentElement.setAttribute('dir', direction)
  document.documentElement.setAttribute('lang', language)
}

void i18next
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    lng: 'en',
    fallbackLng: 'en',
    supportedLngs: ['en', 'ar'],
    interpolation: { escapeValue: false },
    defaultNS: 'common',
    ns: ['common', 'nav', 'pages'],
    resources: {
      en: {
        common: commonEn,
        nav: navEn,
        pages: pagesEn,
      },
      ar: {
        common: commonAr,
      },
    },
    detection: {
      order: ['localStorage', 'navigator'],
      caches: ['localStorage'],
    },
  })
  .then(() => {
    applyDocumentLocale(i18next.resolvedLanguage || i18next.language || 'en')
  })

i18next.on('languageChanged', (language) => {
  applyDocumentLocale(language)
})

export default i18next
