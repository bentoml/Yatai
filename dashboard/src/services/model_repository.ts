import axios from 'axios'
import {
    IModelRepositorySchema,
    ICreateModelRepositorySchema,
    IUpdateModelRepositorySchema,
} from '@/schemas/model_repository'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listModelRepositories(query: IListQuerySchema): Promise<IListSchema<IModelRepositorySchema>> {
    const response = await axios.get<IListSchema<IModelRepositorySchema>>('/api/v1/model_repositories', {
        params: query,
    })
    return response.data
}

export async function fetchModelRepository(modelRepositoryName: string): Promise<IModelRepositorySchema> {
    const response = await axios.get<IModelRepositorySchema>(`/api/v1/model_repositories/${modelRepositoryName}`)
    return response.data
}

export async function createModelRepository(data: ICreateModelRepositorySchema): Promise<IModelRepositorySchema> {
    const response = await axios.post<IModelRepositorySchema>('/api/v1/model_repositories', data)
    return response.data
}

export async function updateModelRepository(
    modelRepositoryName: string,
    data: IUpdateModelRepositorySchema
): Promise<IModelRepositorySchema> {
    const response = await axios.patch<IModelRepositorySchema>(
        `/api/v1/model_repositories/${modelRepositoryName}`,
        data
    )
    return response.data
}
