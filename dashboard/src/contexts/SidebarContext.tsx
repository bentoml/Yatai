import React from 'react'

export interface ISidebarContextProps {
    expanded: boolean
    setExpanded: (expanded: boolean) => void
}

export const SidebarContext = React.createContext<ISidebarContextProps>({
    expanded: true,
    // eslint-disable-next-line @typescript-eslint/no-empty-function
    setExpanded: () => {},
})
