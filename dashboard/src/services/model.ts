import axios from 'axios'
import { IModelSchema, ICreateModelSchema, IUpdateModelSchema } from '@/schemas/model'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listModels(orgName: string, query: IListQuerySchema): Promise<IListSchema<IModelSchema>> {
    const response = await axios.get<IListSchema<IModelSchema>>(`/api/v1/orgs/${orgName}/models`, { params: query })
    return response.data
}

export async function fetchModel(orgName: string, modelName: string): Promise<IModelSchema> {
    const response = await axios.get<IModelSchema>(`/api/v1/orgs/${orgName}/models/${modelName}`)
    return response.data
}

export async function createModel(orgName: string, data: ICreateModelSchema): Promise<IModelSchema> {
    const response = await axios.post<IModelSchema>(`/api/v1/orgs/${orgName}/models`, data)
    return response.data
}

export async function updateModel(
    orgName: string,
    modelName: string,
    model: IUpdateModelSchema
): Promise<IModelSchema> {
    const response = await axios.patch<IModelSchema>(`/api/v1/orgs/${orgName}/models/${modelName}`, model)
    return response.data
}
