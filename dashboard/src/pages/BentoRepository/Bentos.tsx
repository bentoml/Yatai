import React from 'react'
import { useParams } from 'react-router-dom'
import BentoListCard from '@/components/BentoListCard'

export default function BentoRepositoryBentos() {
    const { bentoRepositoryName } = useParams<{ bentoRepositoryName: string }>()

    return <BentoListCard bentoRepositoryName={bentoRepositoryName} />
}
