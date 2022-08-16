import { IBentoFullSchema } from '@/schemas/bento'
import { fetchBento } from '@/services/bento'
import { useQuery } from 'react-query'
import { useOrganization } from './useOrganization'

export function useFetchBento(bentoRepositoryName: string, version: string) {
    const { organization } = useOrganization()
    const bentoInfo = useQuery(`fetchBento:${organization?.name}:${bentoRepositoryName}:${version}`, () =>
        fetchBento(bentoRepositoryName, version)
    )
    return bentoInfo
}

export function useFetchBentoOptional(bentoRepositoryName?: string, version?: string) {
    const { organization } = useOrganization()
    const bentoInfo = useQuery(
        `fetchBentoOptional:${organization?.name}:${bentoRepositoryName}:${version}`,
        (): Promise<IBentoFullSchema | undefined> =>
            bentoRepositoryName && version ? fetchBento(bentoRepositoryName, version) : Promise.resolve(undefined)
    )
    return bentoInfo
}
