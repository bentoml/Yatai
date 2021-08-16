import React from 'react'
import BaseLayout from './BaseLayout'

export interface IYataiLayoutProps {
    children: React.ReactNode
}

export default function YataiLayout({ children }: IYataiLayoutProps) {
    return <BaseLayout>{children}</BaseLayout>
}
