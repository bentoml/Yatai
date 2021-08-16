import { headerHeight } from '@/consts'
import React from 'react'
import { IComposedSidebarProps } from './BaseSidebar'

export interface IBaseLayoutProps {
    children: React.ReactNode
    sidebar?: React.ComponentType<IComposedSidebarProps>
    contentStyle?: React.CSSProperties
}

export default function BaseLayout({ children, sidebar: Sidebar, contentStyle }: IBaseLayoutProps) {
    return (
        <main
            style={{
                height: '100vh',
                display: 'flex',
                flexFlow: 'row nowrap',
                justifyContent: 'space-between',
            }}
        >
            {Sidebar && <Sidebar style={{ marginTop: headerHeight }} />}
            <div
                style={{
                    overflowY: 'auto',
                    paddingTop: headerHeight,
                    flexGrow: 1,
                }}
            >
                <div
                    style={{
                        padding: '20px',
                        height: '100%',
                        boxSizing: 'border-box',
                        ...contentStyle,
                    }}
                >
                    {children}
                </div>
            </div>
        </main>
    )
}
