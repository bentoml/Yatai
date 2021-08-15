import React from 'react'
import BaseLayout from './BaseLayout'

export interface IBundleLayoutProps {
    children: React.ReactNode
}

export default function BundleLayout({ children }: IBundleLayoutProps) {
    return <BaseLayout>{children}</BaseLayout>
}
