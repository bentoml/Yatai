import React from 'react'
import { Client as Styletron } from 'styletron-engine-atomic'
import { Provider as StyletronProvider } from 'styletron-react'
import { LightTheme, BaseProvider, DarkTheme } from 'baseui'
import { ToasterContainer } from 'baseui/toast'
import { SidebarContext } from '@/contexts/SidebarContext'
import { useSidebar } from '@/hooks/useSidebar'
import Routes from '@/routes'
import { QueryClient, QueryClientProvider } from 'react-query'
import { useCurrentThemeType } from './hooks/useCurrentThemeType'

const engine = new Styletron()
const queryClient = new QueryClient()

export default function Hello() {
    const sidebarData = useSidebar()
    const themeType = useCurrentThemeType()

    return (
        <QueryClientProvider client={queryClient}>
            <StyletronProvider value={engine}>
                <BaseProvider theme={themeType === 'dark' ? DarkTheme : LightTheme}>
                    <ToasterContainer>
                        <SidebarContext.Provider value={sidebarData}>
                            <Routes />
                        </SidebarContext.Provider>
                    </ToasterContainer>
                </BaseProvider>
            </StyletronProvider>
        </QueryClientProvider>
    )
}
