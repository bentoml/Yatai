import React from 'react'
import BaseLayout from './BaseLayout'
import ClusterSidebar from './ClusterSidebar'

export interface IClusterLayoutProps {
    children: React.ReactNode
}

export default function ClusterLayout({ children }: IClusterLayoutProps) {
    return <BaseLayout sidebar={ClusterSidebar}>{children}</BaseLayout>
}
