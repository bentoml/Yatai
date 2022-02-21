import DeploymentForm from '@/components/DeploymentForm'
import { useCallback } from 'react'
import { IUpdateDeploymentSchema } from '@/schemas/deployment'
import { useParams } from 'react-router-dom'
import { updateDeployment } from '@/services/deployment'
import { useFetchDeployment } from '@/hooks/useFetchDeployment'
import { Skeleton } from 'baseui/skeleton'

export default function DeploymentEdit() {
    const { clusterName, kubeNamespace, deploymentName } =
        useParams<{ clusterName: string; kubeNamespace: string; deploymentName: string }>()
    const { deploymentInfo } = useFetchDeployment(clusterName, kubeNamespace, deploymentName)
    const handleCreateDeploymentRevision = useCallback(
        async (data: IUpdateDeploymentSchema) => {
            await updateDeployment(clusterName, kubeNamespace, deploymentName, data)
            deploymentInfo.refetch()
        },
        [clusterName, deploymentInfo, deploymentName, kubeNamespace]
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
