import { IListQuerySchema } from '@/schemas/list'
import { useQuery } from 'react-query'
import { listApiTokens } from '@/services/api_token'
import qs from 'qs'
import { useOrganization } from './useOrganization'

export function useFetchApiTokens(query: IListQuerySchema) {
    const { organization } = useOrganization()
    const apiTokensInfo = useQuery(`fetchOrgApiTokens:${organization?.name}:${qs.stringify(query)}`, () =>
        listApiTokens(query)
    )
    return apiTokensInfo
}
