import { IListQuerySchema } from '@/schemas/list'
import { useQuery } from 'react-query'
import { listApiTokens } from '@/services/api_token'

export function useFetchApiTokens(query: IListQuerySchema) {
    const apiTokensInfo = useQuery('fetchOrgApiTokens', () => listApiTokens(query))
    return apiTokensInfo
}
