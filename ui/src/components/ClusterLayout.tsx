import React from 'react'
import BaseLayout from './BaseLayout'

export interface IClusterLayoutProps {
    children: React.ReactNode
}

export default function ClusterLayout({ children }: IClusterLayoutProps) {
    return <BaseLayout>{children}</BaseLayout>
}
