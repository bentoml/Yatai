import axios from 'axios'
import { IModelSchema, ICreateModelSchema, IUpdateModelSchema } from '@/schemas/model'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listModels(query: IListQuerySchema): Promise<IListSchema<IModelSchema>> {
    const response = await axios.get<IListSchema<IModelSchema>>('/api/v1/models', { params: query })
    return response.data
}

export async function fetchModel(modelName: string): Promise<IModelSchema> {
    const response = await axios.get<IModelSchema>(`/api/v1/models/${modelName}`)
    return response.data
}

export async function createModel(data: ICreateModelSchema): Promise<IModelSchema> {
    const response = await axios.post<IModelSchema>('/api/v1/models', data)
    return response.data
}

export async function updateModel(modelName: string, model: IUpdateModelSchema): Promise<IModelSchema> {
    const response = await axios.patch<IModelSchema>(`/api/v1/models/${modelName}`, model)
    return response.data
}
