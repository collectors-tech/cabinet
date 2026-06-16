import commonEn from '@/locales/en/common.json'
import navEn from '@/locales/en/nav.json'
import pagesEn from '@/locales/en/pages.json'
import commonJa from '@/locales/ja/common.json'
import navJa from '@/locales/ja/nav.json'
import pagesJa from '@/locales/ja/pages.json'
import commonZh from '@/locales/zh/common.json'
import navZh from '@/locales/zh/nav.json'
import pagesZh from '@/locales/zh/pages.json'
import i18next from 'i18next'
import LanguageDetector from 'i18next-browser-languagedetector'
import { initReactI18next } from 'react-i18next'

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
    supportedLngs: ['en', 'zh', 'ja'],
    interpolation: { escapeValue: false },
    defaultNS: 'common',
    ns: ['common', 'nav', 'pages'],
    resources: {
      en: {
        common: commonEn,
        nav: navEn,
        pages: pagesEn,
      },
      zh: {
        common: commonZh,
        nav: navZh,
        pages: pagesZh,
      },
      ja: {
        common: commonJa,
        nav: navJa,
        pages: pagesJa,
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
