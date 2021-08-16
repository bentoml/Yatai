import { useCallback, useEffect, useState } from 'react'

export const useThemeDetector = () => {
    const getCurrentTheme = () => window.matchMedia('(prefers-color-scheme: dark)').matches
    const [isDarkTheme, setIsDarkTheme] = useState(getCurrentTheme())
    const mqListener = useCallback((e) => {
        setIsDarkTheme(e.matches)
    }, [])

    useEffect(() => {
        const darkThemeMq = window.matchMedia('(prefers-color-scheme: dark)')
        darkThemeMq.addEventListener('change', mqListener)
        return () => darkThemeMq.removeEventListener('change', mqListener)
    }, [mqListener])

    return isDarkTheme
}
