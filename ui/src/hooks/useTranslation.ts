import { useCallback } from 'react'
import { useTranslation as _useTranslation } from 'react-i18next'
import * as i18next from 'i18next'
import { locales } from '@/i18n/locales'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type Translator = (key: keyof typeof locales, options?: { [key: string]: any }) => string

export type UseTranslationResponse = [Translator, i18next.i18n]

export default function useTranslation(): UseTranslationResponse {
    const [t0, i18n] = _useTranslation()

    const t = useCallback(
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        (key: keyof typeof locales, options?: { [key: string]: any }): string => {
            return t0(key as string, options)
        },
        [t0]
    )

    return [t, i18n]
}
