import ModelVersionListCard from '@/components/ModelVersionListCard'
import React from 'react'
import { useParams } from 'react-router-dom'

export default function ModelVersions() {
    const { modelName } = useParams<{ modelName: string }>()

    return <ModelVersionListCard modelName={modelName} />
}
