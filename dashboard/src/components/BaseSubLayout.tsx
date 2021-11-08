import React from 'react'
import { INavItem } from './BaseSidebar'
import { BaseNavTabs } from './BaseNavTabs'
import BaseLayout from './BaseLayout'
import OrganizationSidebar from './OrganizationSidebar'
import Card from './Card'

export interface IBaseSubLayoutProps {
    header?: React.ReactNode
    breadcrumbItems?: INavItem[]
    navItems: INavItem[]
    children: React.ReactNode
}

export default function BaseSubLayout({ header, breadcrumbItems, navItems, children }: IBaseSubLayoutProps) {
    return (
        <BaseLayout breadcrumbItems={breadcrumbItems} sidebar={OrganizationSidebar}>
            {header}
            <Card bodyStyle={{ padding: 0 }}>
                <BaseNavTabs navItems={navItems} />
                <div
                    style={{
                        paddingTop: 5,
                        paddingBottom: 15,
                    }}
                >
                    {children}
                </div>
            </Card>
        </BaseLayout>
    )
}
