import ModelListCard from '@/components/ModelListCard'
import React from 'react'
import { useParams } from 'react-router-dom'

export default function OrganizationModels() {
    const { orgName } = useParams<{ orgName: string }>()

    return <ModelListCard orgName={orgName} />
}
