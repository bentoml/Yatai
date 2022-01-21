import DeploymentForm from '@/components/DeploymentForm'
import { useCallback } from 'react'
import { IUpdateDeploymentSchema } from '@/schemas/deployment'
import { useParams } from 'react-router-dom'
import { updateDeployment } from '@/services/deployment'
import { useFetchDeployment } from '@/hooks/useFetchDeployment'
import { Skeleton } from 'baseui/skeleton'

export default function DeploymentEdit() {
    const { clusterName, deploymentName } = useParams<{ clusterName: string; deploymentName: string }>()
    const { deploymentInfo } = useFetchDeployment(clusterName, deploymentName)
    const handleCreateDeploymentRevision = useCallback(
        async (data: IUpdateDeploymentSchema) => {
            await updateDeployment(clusterName, deploymentName, data)
            deploymentInfo.refetch()
        },
        [clusterName, deploymentInfo, deploymentName]
    )

    if (deploymentInfo.isLoading) {
        return <Skeleton animation rows={3} />
    }

    return (
        <DeploymentForm
            deployment={deploymentInfo.data}
            deploymentRevision={deploymentInfo.data?.latest_revision}
            onSubmit={handleCreateDeploymentRevision}
        />
    )
}
