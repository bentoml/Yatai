import { useQuery } from 'react-query'
import { IListQuerySchema } from '@/schemas/list'
import { listDeploymentRevisions } from '@/services/deployment_revision'
import qs from 'qs'

export function useFetchDeploymentRevisions(clusterName: string, deploymentName: string, page: IListQuerySchema) {
    const queryKey = `fetchDeploymentRevisions:${clusterName}:${deploymentName}:${qs.stringify(page)}`
    const deploymentRevisionsInfo = useQuery(queryKey, () => listDeploymentRevisions(clusterName, deploymentName, page))
    return { deploymentRevisionsInfo }
}
