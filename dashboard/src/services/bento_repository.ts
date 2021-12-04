import axios from 'axios'
import {
    ICreateBentoRepositorySchema,
    IBentoRepositorySchema,
    IUpdateBentoRepositorySchema,
} from '@/schemas/bento_repository'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listBentoRepositories(query: IListQuerySchema): Promise<IListSchema<IBentoRepositorySchema>> {
    const resp = await axios.get<IListSchema<IBentoRepositorySchema>>('/api/v1/bento_repositories', { params: query })
    return resp.data
}

export async function fetchBentoRepository(bentoRepositoryName: string): Promise<IBentoRepositorySchema> {
    const resp = await axios.get<IBentoRepositorySchema>(`/api/v1/bento_repositories/${bentoRepositoryName}`)
    return resp.data
}

export async function createBentoRepository(data: ICreateBentoRepositorySchema): Promise<IBentoRepositorySchema> {
    const resp = await axios.post<IBentoRepositorySchema>('/api/v1/bento_repositories', data)
    return resp.data
}

export async function updateBentoRepository(
    bentoRepositoryName: string,
    data: IUpdateBentoRepositorySchema
): Promise<IBentoRepositorySchema> {
    const resp = await axios.patch<IBentoRepositorySchema>(`/api/v1/bento_repositories/${bentoRepositoryName}`, data)
    return resp.data
}
