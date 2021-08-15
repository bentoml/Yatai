import { ISidebarContextProps } from '@/contexts/SidebarContext'
import { useState, useCallback } from 'react'

export const useSidebar = (): ISidebarContextProps => {
    const [expanded, _setExpanded] = useState(true)

    const setExpanded = useCallback((expanded_: boolean) => {
        _setExpanded(expanded_)
    }, [])

    return {
        expanded,
        setExpanded,
    }
}
