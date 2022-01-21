import DeploymentForm from '@/components/DeploymentForm'
import { ICreateDeploymentSchema } from '@/schemas/deployment'
import { createDeployment } from '@/services/deployment'
import React, { useCallback } from 'react'

export default function OrganizationDeploymentForm() {
    const handleCreateDeployment = useCallback(async (data: ICreateDeploymentSchema) => {
        if (!data.cluster_name) {
            return
        }
        await createDeployment(data.cluster_name, data)
    }, [])
    return <DeploymentForm onSubmit={handleCreateDeployment} />
}
