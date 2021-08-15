import { IListQuerySchema } from '@/schemas/list'
import { useQuery } from 'react-query'
import { listClusters } from '@/services/cluster'

export function useFetchClusters(orgName: string | undefined = '', query: IListQuerySchema) {
    const clustersInfo = useQuery(`fetchOrgClusters:${orgName}`, () => listClusters(orgName, query))
    return clustersInfo
}
