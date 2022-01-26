import { fetchBento } from '@/services/bento'
import { useQuery } from 'react-query'

export function useFetchBento(bentoRepositoryName: string, version: string) {
    const bentoInfo = useQuery(`fetchBento:${bentoRepositoryName}:${version}`, () =>
        fetchBento(bentoRepositoryName, version)
    )
    return bentoInfo
}
