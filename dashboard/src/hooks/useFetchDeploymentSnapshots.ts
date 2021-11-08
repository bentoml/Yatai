import { useQuery } from 'react-query'
import { IListQuerySchema } from '@/schemas/list'
import { listDeploymentSnapshots } from '@/services/deployment_snapshot'

export function useFetchDeploymentSnapshots(clusterName: string, deploymentName: string, page: IListQuerySchema) {
    const queryKey = `fetchDeploymentSnapshots:${clusterName}:${deploymentName}:${page.start}:${page.count}`
    const deploymentSnapshotsInfo = useQuery(queryKey, () => listDeploymentSnapshots(clusterName, deploymentName, page))
    return { deploymentSnapshotsInfo }
}
