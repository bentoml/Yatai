import { fetchDeployment } from '@/services/deployment'
import { useQuery } from 'react-query'
import { useOrganization } from './useOrganization'

export function useFetchDeployment(clusterName: string, kubeNamespace: string, deploymentName: string) {
    const { organization } = useOrganization()
    const queryKey = `fetchDeployment:${organization?.name}:${clusterName}:${kubeNamespace}:${deploymentName}`
    const deploymentInfo = useQuery(queryKey, () => fetchDeployment(clusterName, kubeNamespace, deploymentName))
    return { queryKey, deploymentInfo }
}
