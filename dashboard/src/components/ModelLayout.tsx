import React from 'react'
import BaseLayout from './BaseLayout'

export interface IModelLayoutProps {
    children: React.ReactNode
}

export default function ModelLayout({ children }: IModelLayoutProps) {
    return <BaseLayout>{children}</BaseLayout>
}
