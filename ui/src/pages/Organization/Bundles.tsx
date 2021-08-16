import React from 'react'
import { useParams } from 'react-router-dom'
import BundleListCard from '@/components/BundleListCard'

export default function OrganizationBundles() {
    const { orgName } = useParams<{ orgName: string }>()

    return <BundleListCard orgName={orgName} />
}
