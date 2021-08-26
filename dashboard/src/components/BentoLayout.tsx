import React from 'react'
import BaseLayout from './BaseLayout'
import BentoSidebar from './BentoSidebar'

export interface IBentoLayoutProps {
    children: React.ReactNode
}

export default function BentoLayout({ children }: IBentoLayoutProps) {
    return <BaseLayout sidebar={BentoSidebar}>{children}</BaseLayout>
}
