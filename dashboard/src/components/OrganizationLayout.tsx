import React from 'react'
import BaseLayout from './BaseLayout'
import OrganizationSidebar from './OrganizationSidebar'

export interface IOrganizationLayoutProps {
    children: React.ReactNode
}

export default function OrganizationLayout({ children }: IOrganizationLayoutProps) {
    return <BaseLayout sidebar={OrganizationSidebar}>{children}</BaseLayout>
}
