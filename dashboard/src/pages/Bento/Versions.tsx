import React from 'react'
import { useParams } from 'react-router-dom'
import BentoVersionListCard from '@/components/BentoVersionListCard'

export default function BentoVersions() {
    const { bentoName } = useParams<{ bentoName: string }>()

    return <BentoVersionListCard bentoName={bentoName} />
}
