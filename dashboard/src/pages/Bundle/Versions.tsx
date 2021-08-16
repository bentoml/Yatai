import React from 'react'
import { useParams } from 'react-router-dom'
import BundleVersionListCard from '@/components/BundleVersionListCard'

export default function BundleVersions() {
    const { orgName, bundleName } = useParams<{ orgName: string; bundleName: string }>()

    return <BundleVersionListCard orgName={orgName} bundleName={bundleName} />
}
