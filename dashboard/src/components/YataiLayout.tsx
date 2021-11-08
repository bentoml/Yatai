import React from 'react'
import BaseLayout from './BaseLayout'

export interface IYataiLayoutProps {
    children: React.ReactNode
    style?: React.CSSProperties
}

export default function YataiLayout({ children, style }: IYataiLayoutProps) {
    return <BaseLayout style={style}>{children}</BaseLayout>
}
