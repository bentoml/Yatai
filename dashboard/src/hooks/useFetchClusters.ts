import { IListQuerySchema } from '@/schemas/list'
import { useQuery } from 'react-query'
import { listClusters } from '@/services/cluster'

export function useFetchClusters(query: IListQuerySchema) {
    const clustersInfo = useQuery('fetchOrgClusters', () => listClusters(query))
    return clustersInfo
}
