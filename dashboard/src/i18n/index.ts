import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import { locales } from '@/i18n/locales'
import LanguageDetector from 'i18next-browser-languagedetector'
import TimeAgo from 'javascript-time-ago'

import en from 'javascript-time-ago/locale/en.json'
import zh from 'javascript-time-ago/locale/zh.json'
import ja from 'javascript-time-ago/locale/ja.json'
import ko from 'javascript-time-ago/locale/ko.json'
import vi from 'javascript-time-ago/locale/vi.json'

TimeAgo.addDefaultLocale(en)
TimeAgo.addLocale(zh)
TimeAgo.addLocale(ja)
TimeAgo.addLocale(ko)
TimeAgo.addLocale(vi)

i18n.use(LanguageDetector)
    .use(initReactI18next)
    .init({
        // we init with resources
        resources: {
            en: {
                translations: Object.entries(locales).reduce((p, [k, v]) => {
                    return {
                        ...p,
                        [k]: v.en,
                    }
                }, {}),
            },
            zh: {
                translations: Object.entries(locales).reduce((p, [k, v]) => {
                    return {
                        ...p,
                        [k]: v.zh,
                    }
                }, {}),
            },
            ja: {
                translations: Object.entries(locales).reduce((p, [k, v]) => {
                    return {
                        ...p,
                        [k]: v.ja,
                    }
                }, {}),
            },
            ko: {
                translations: Object.entries(locales).reduce((p, [k, v]) => {
                    return {
                        ...p,
                        [k]: v.ko,
                    }
                }, {}),
            },
            vi: {
                translations: Object.entries(locales).reduce((p, [k, v]) => {
                    return {
                        ...p,
                        [k]: v.vi,
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
