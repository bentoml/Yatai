import { IListQuerySchema } from '@/schemas/list'
import { useQuery } from 'react-query'
import { listClusters } from '@/services/cluster'
import qs from 'qs'
import { useOrganization } from './useOrganization'

export function useFetchClusters(query: IListQuerySchema) {
    const { organization } = useOrganization()
    const clustersInfo = useQuery(`fetchOrgClusters:${organization?.name}:${qs.stringify(query)}`, () =>
        listClusters(query)
    )
    return clustersInfo
}
