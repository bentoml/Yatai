import React from 'react'
import { useParams } from 'react-router-dom'
import ClusterMemberListCard from '@/components/ClusterMemberListCard'

export default function ClusterMembers() {
    const { orgName, clusterName } = useParams<{ orgName: string; clusterName: string }>()

    return <ClusterMemberListCard orgName={orgName} clusterName={clusterName} />
}
