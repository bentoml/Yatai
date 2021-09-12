import ModelVersionListCard from '@/components/ModelVersionListCard'
import React from 'react'
import { useParams } from 'react-router-dom'

export default function ModelVersions() {
    const { orgName, modelName } = useParams<{ orgName: string; modelName: string }>()

    return <ModelVersionListCard orgName={orgName} modelName={modelName} />
}
