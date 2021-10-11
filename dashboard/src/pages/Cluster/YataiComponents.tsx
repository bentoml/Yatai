import React from 'react'
import { useParams } from 'react-router-dom'
import YataiListCard from '@/components/YataiComponentListCard'

export default function ClusterDeployments() {
    const { orgName, clusterName } = useParams<{ orgName: string; clusterName: string }>()

    return <YataiListCard orgName={orgName} clusterName={clusterName} />
}
