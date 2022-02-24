import { fetchDeployment } from '@/services/deployment'
import { useQuery } from 'react-query'

export function useFetchDeployment(clusterName: string, kubeNamespace: string, deploymentName: string) {
    const queryKey = `fetchDeployment:${clusterName}:${kubeNamespace}:${deploymentName}`
    const deploymentInfo = useQuery(queryKey, () => fetchDeployment(clusterName, kubeNamespace, deploymentName))
    return { queryKey, deploymentInfo }
}
