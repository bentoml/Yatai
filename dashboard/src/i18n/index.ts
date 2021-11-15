import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import { locales } from '@/i18n/locales'
import LanguageDetector from 'i18next-browser-languagedetector'

i18n.use(LanguageDetector)
    .use(initReactI18next)
    .init({
        // we init with resources
        resources: {
            'en': {
                translations: Object.entries(locales).reduce((p, [k, v]) => {
                    return {
                        ...p,
                        [k]: v.en,
                    }
                }, {}),
            },
            'zh-CN': {
                translations: Object.entries(locales).reduce((p, [k, v]) => {
                    return {
                        ...p,
                        [k]: v.cn,
                    }
                }, {}),
            },
            'ja': {
                translations: Object.entries(locales).reduce((p, [k, v]) => {
                    return {
                        ...p,
                        [k]: v.ja,
                    }
                }, {}),
            },
        },
        fallbackLng: 'en',
        debug: false,

        // have a common namespace used around the full app
        ns: ['translations'],
        defaultNS: 'translations',

        keySeparator: false, // we use content as keys

        interpolation: {
            escapeValue: false,
        },
    })

export default i18n
