import React from 'react'
import { INavItem } from './BaseSidebar'
import { BaseNavTabs } from './BaseNavTabs'
import BaseLayout from './BaseLayout'
import OrganizationSidebar from './OrganizationSidebar'
import Card from './Card'

export interface IBaseSubLayoutProps {
    header?: React.ReactNode
    extra?: React.ReactNode
    breadcrumbItems?: INavItem[]
    navItems?: INavItem[]
    children: React.ReactNode
}

export default function BaseSubLayout({ header, extra, breadcrumbItems, navItems, children }: IBaseSubLayoutProps) {
    return (
        <BaseLayout extra={extra} breadcrumbItems={breadcrumbItems} sidebar={OrganizationSidebar}>
            {header}
            {navItems ? (
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
            ) : (
                <div>{children}</div>
            )}
        </BaseLayout>
    )
}
