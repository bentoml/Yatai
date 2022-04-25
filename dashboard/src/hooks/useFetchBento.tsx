import { IBentoFullSchema } from '@/schemas/bento'
import { fetchBento } from '@/services/bento'
import { useQuery } from 'react-query'

export function useFetchBento(bentoRepositoryName: string, version: string) {
    const bentoInfo = useQuery(`fetchBento:${bentoRepositoryName}:${version}`, () =>
        fetchBento(bentoRepositoryName, version)
    )
    return bentoInfo
}

export function useFetchBentoOptional(bentoRepositoryName?: string, version?: string) {
    const bentoInfo = useQuery(
        `fetchBentoOptional:${bentoRepositoryName}:${version}`,
        (): Promise<IBentoFullSchema | undefined> =>
            bentoRepositoryName && version ? fetchBento(bentoRepositoryName, version) : Promise.resolve(undefined)
    )
    return bentoInfo
}
