import React from 'react'
import { useParams } from 'react-router-dom'
import ClusterListCard from '@/components/ClusterListCard'

export default function OrganizationClusters() {
    const { orgName } = useParams<{ orgName: string }>()

    return <ClusterListCard orgName={orgName} />
}
