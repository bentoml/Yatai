import React from 'react'
import BaseLayout from './BaseLayout'
import BundleSidebar from './BundleSidebar'

export interface IBundleLayoutProps {
    children: React.ReactNode
}

export default function BundleLayout({ children }: IBundleLayoutProps) {
    return <BaseLayout sidebar={BundleSidebar}>{children}</BaseLayout>
}
