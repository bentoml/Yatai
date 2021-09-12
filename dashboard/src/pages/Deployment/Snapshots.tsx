import React from 'react'
import { useParams } from 'react-router-dom'
import DeploymentSnapshotListCard from '@/components/DeploymentSnapshotListCard'

export default function DeploymentSnapshots() {
    const { orgName, clusterName, deploymentName } =
        useParams<{ orgName: string; clusterName: string; deploymentName: string }>()

    return <DeploymentSnapshotListCard orgName={orgName} clusterName={clusterName} deploymentName={deploymentName} />
}
