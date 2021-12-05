import React from 'react'
import BaseLayout from './BaseLayout'

export interface IModelRepositoryLayoutProps {
    children: React.ReactNode
}

export default function ModelRepositoryLayout({ children }: IModelRepositoryLayoutProps) {
    return <BaseLayout>{children}</BaseLayout>
}
