import { useQuery } from 'react-query'
import { fetchCluster } from '@/services/cluster'

export function useFetchCluster(name: string) {
    const clusterInfo = useQuery(`fetchOrgCluster:${name}`, () => fetchCluster(name))
    return clusterInfo
}
