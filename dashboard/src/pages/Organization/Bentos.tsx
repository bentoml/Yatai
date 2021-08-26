import React from 'react'
import { useParams } from 'react-router-dom'
import BentoListCard from '@/components/BentoListCard'

export default function OrganizationBentos() {
    const { orgName } = useParams<{ orgName: string }>()

    return <BentoListCard orgName={orgName} />
}
