import { IListQuerySchema } from '@/schemas/list'
import { useQuery } from 'react-query'
import { listClusters } from '@/services/cluster'
import qs from 'qs'

export function useFetchClusters(query: IListQuerySchema) {
    const clustersInfo = useQuery(`fetchOrgClusters:${qs.stringify(query)}`, () => listClusters(query))
    return clustersInfo
}
