import { fetchCurrentUserApiToken } from '@/services/user'
import { useQuery } from 'react-query'

export function useFetchCurrentUserApiToken() {
    const apiInfo = useQuery('fetchCurrentUserApiToken', () => fetchCurrentUserApiToken())
    return apiInfo
}
