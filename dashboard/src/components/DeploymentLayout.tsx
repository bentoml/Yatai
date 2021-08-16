import React from 'react'
import BaseLayout from './BaseLayout'

export interface IDeploymentLayoutProps {
    children: React.ReactNode
}

export default function DeploymentLayout({ children }: IDeploymentLayoutProps) {
    return <BaseLayout>{children}</BaseLayout>
}
