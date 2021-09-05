import React from 'react'
import BaseLayout from './BaseLayout'
import DeploymentSidebar from './DeploymentSidebar'

export interface IDeploymentLayoutProps {
    children: React.ReactNode
}

export default function DeploymentLayout({ children }: IDeploymentLayoutProps) {
    return <BaseLayout sidebar={DeploymentSidebar}>{children}</BaseLayout>
}
