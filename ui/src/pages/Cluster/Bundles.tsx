import React from 'react'
import { useParams } from 'react-router-dom'
import BundleListCard from '@/components/BundleListCard'

export default function ClusterBundles() {
    const { orgName, clusterName } = useParams<{ orgName: string; clusterName: string }>()

    return <BundleListCard orgName={orgName} clusterName={clusterName} />
}
