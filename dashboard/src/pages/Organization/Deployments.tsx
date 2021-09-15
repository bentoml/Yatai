import React from 'react'
import { useParams } from 'react-router-dom'
import DeploymentListCard from '@/components/DeploymentListCard'

export default function OrganizationDeployments() {
    const { orgName } = useParams<{ orgName: string }>()

    return <DeploymentListCard orgName={orgName} />
}
