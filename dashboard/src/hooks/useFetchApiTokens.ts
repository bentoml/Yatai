import { IListQuerySchema } from '@/schemas/list'
import { useQuery } from 'react-query'
import { listApiTokens } from '@/services/api_token'
import qs from 'qs'

export function useFetchApiTokens(query: IListQuerySchema) {
    const apiTokensInfo = useQuery(`fetchOrgApiTokens:${qs.stringify(query)}`, () => listApiTokens(query))
    return apiTokensInfo
}
