import React from 'react'
import { useParams } from 'react-router-dom'
import DeploymentListCard from '@/components/DeploymentListCard'

export default function ClusterDeployments() {
    const { orgName, clusterName } = useParams<{ orgName: string; clusterName: string }>()

    return <DeploymentListCard orgName={orgName} clusterName={clusterName} />
}
