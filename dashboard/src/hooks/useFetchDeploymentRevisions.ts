import { useQuery } from 'react-query'
import { IListQuerySchema } from '@/schemas/list'
import { listDeploymentRevisions } from '@/services/deployment_revision'

export function useFetchDeploymentRevisions(clusterName: string, deploymentName: string, page: IListQuerySchema) {
    const queryKey = `fetchDeploymentRevisions:${clusterName}:${deploymentName}:${page.start}:${page.count}`
    const deploymentRevisionsInfo = useQuery(queryKey, () => listDeploymentRevisions(clusterName, deploymentName, page))
    return { deploymentRevisionsInfo }
}
