import React from 'react'
import { useParams } from 'react-router-dom'
import OrganizationMemberListCard from '@/components/OrganizationMemberListCard'

export default function OrganizationMembers() {
    const { orgName } = useParams<{ orgName: string }>()

    return <OrganizationMemberListCard orgName={orgName} />
}
