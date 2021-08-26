import React from 'react'
import { useParams } from 'react-router-dom'
import BentoVersionListCard from '@/components/BentoVersionListCard'

export default function BentoVersions() {
    const { orgName, bentoName } = useParams<{ orgName: string; bentoName: string }>()

    return <BentoVersionListCard orgName={orgName} bentoName={bentoName} />
}
