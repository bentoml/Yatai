import DeploymentForm from '@/components/DeploymentForm'
import { useCallback } from 'react'
import { IUpdateDeploymentSchema } from '@/schemas/deployment'
import { useParams } from 'react-router-dom'
import { updateDeployment } from '@/services/deployment'
import { useFetchDeployment } from '@/hooks/useFetchDeployment'
import { Skeleton } from 'baseui/skeleton'
import { fetchDeploymentRevision } from '@/services/deployment_revision'
import { useQuery } from 'react-query'

export default function DeploymentRevisionRollback() {
    const { clusterName, kubeNamespace, deploymentName, revisionUid } =
        useParams<{ clusterName: string; kubeNamespace: string; deploymentName: string; revisionUid: string }>()
    const { deploymentInfo } = useFetchDeployment(clusterName, kubeNamespace, deploymentName)
    const deploymentRevisionInfo = useQuery(
        `fetchDeploymentRevision:${clusterName}:${kubeNamespace}:${deploymentName}:${revisionUid}`,
        () => fetchDeploymentRevision(clusterName, kubeNamespace, deploymentName, revisionUid)
    )
    const handleCreateDeploymentRevision = useCallback(
        async (data: IUpdateDeploymentSchema) => {
            await updateDeployment(clusterName, kubeNamespace, deploymentName, data)
            deploymentInfo.refetch()
        },
        [clusterName, deploymentInfo, deploymentName, kubeNamespace]
    )

    if (deploymentInfo.isLoading || deploymentRevisionInfo.isLoading) {
        return <Skeleton animation rows={3} />
    }

    return (
        <DeploymentForm
            deployment={deploymentInfo.data}
            deploymentRevision={deploymentRevisionInfo.data}
            onSubmit={handleCreateDeploymentRevision}
        />
    )
}
