import ModelListCard from '@/components/ModelListCard'
import React from 'react'
import { useParams } from 'react-router-dom'

export default function ModelRepositoryModels() {
    const { modelRepositoryName } = useParams<{ modelRepositoryName: string }>()

    return <ModelListCard modelRepositoryName={modelRepositoryName} />
}
