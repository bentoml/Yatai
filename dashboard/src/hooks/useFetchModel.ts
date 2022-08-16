import { fetchModel } from '@/services/model'
import { useQuery } from 'react-query'
import { useOrganization } from './useOrganization'

export function useFetchModel(modelRepositoryName: string, version: string) {
    const { organization } = useOrganization()
    const modelInfo = useQuery(`fetchModel:${organization?.name}:${modelRepositoryName}:${version}`, () =>
        fetchModel(modelRepositoryName, version)
    )
    return modelInfo
}
