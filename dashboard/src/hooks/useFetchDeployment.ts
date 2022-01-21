import { fetchDeployment } from '@/services/deployment'
import { useQuery } from 'react-query'

export function useFetchDeployment(clusterName: string, deploymentName: string) {
    const queryKey = `fetchDeployment:${clusterName}:${deploymentName}`
    const deploymentInfo = useQuery(queryKey, () => fetchDeployment(clusterName, deploymentName))
    return { queryKey, deploymentInfo }
}
