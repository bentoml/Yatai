import React from 'react'
import { useParams } from 'react-router-dom'
import YataiListCard from '@/components/YataiComponentListCard'

export default function ClusterDeployments() {
    const { clusterName } = useParams<{ clusterName: string }>()

    return <YataiListCard clusterName={clusterName} />
}
