import i18next from 'i18next'
import LanguageDetector from 'i18next-browser-languagedetector'
import { initReactI18next } from 'react-i18next'
import commonEn from '@/locales/en/common.json'
import navEn from '@/locales/en/nav.json'
import pagesEn from '@/locales/en/pages.json'

void i18next
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    lng: 'en',
    fallbackLng: 'en',
    supportedLngs: ['en'],
    interpolation: { escapeValue: false },
    defaultNS: 'common',
    ns: ['common', 'nav', 'pages'],
    resources: {
      en: {
        common: commonEn,
        nav: navEn,
        pages: pagesEn,
      },
    },
    detection: {
      order: ['localStorage', 'navigator'],
      caches: ['localStorage'],
    },
  })

export default i18next
