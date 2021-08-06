import React from 'react'

export interface ILayoutProps {
    children: React.ReactNode
}

export default function Layout({ children }: ILayoutProps) {
    return <main style={{ padding: '10px 20px' }}>{children}</main>
}
