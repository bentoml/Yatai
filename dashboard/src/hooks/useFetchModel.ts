import { fetchModel } from '@/services/model'
import { useQuery } from 'react-query'

export function useFetchModel(modelRepositoryName: string, version: string) {
    const modelInfo = useQuery(`fetchModel:${modelRepositoryName}:${version}`, () =>
        fetchModel(modelRepositoryName, version)
    )
    return modelInfo
}
