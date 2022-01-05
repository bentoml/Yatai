import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import { locales } from '@/i18n/locales'
import LanguageDetector from 'i18next-browser-languagedetector'
import TimeAgo from 'javascript-time-ago'

import en from 'javascript-time-ago/locale/en.json'
import zh from 'javascript-time-ago/locale/zh.json'
import ja from 'javascript-time-ago/locale/ja.json'

TimeAgo.addDefaultLocale(en)
TimeAgo.addLocale(zh)
TimeAgo.addLocale(ja)

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
            'kr': {
                translations: Object.entries(locales).reduce((p, [k, v]) => {
                    return {
                        ...p,
                        [k]: v.kr,
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
