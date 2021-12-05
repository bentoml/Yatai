import axios from 'axios'

import {
    IModelSchema,
    ICreateModelSchema,
    IFinishedUploadModelSchema,
    IModelWithRepositorySchema,
} from '@/schemas/model'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listAllModels(query: IListQuerySchema): Promise<IListSchema<IModelWithRepositorySchema>> {
    const resp = await axios.get<IListSchema<IModelWithRepositorySchema>>('/api/v1/models', {
        params: query,
    })
    return resp.data
}

export async function listModels(
    modelRepositoryName: string,
    query: IListQuerySchema
): Promise<IListSchema<IModelSchema>> {
    const resp = await axios.get<IListSchema<IModelSchema>>(
        `/api/v1/model_repositories/${modelRepositoryName}/models`,
        {
            params: query,
        }
    )
    return resp.data
}

export async function fetchModel(modelRepositoryName: string, version: string): Promise<IModelSchema> {
    const resp = await axios.get<IModelSchema>(`/api/v1/model_repositories/${modelRepositoryName}/models/${version}`)
    return resp.data
}

export async function createModel(modelRepositoryName: string, data: ICreateModelSchema): Promise<IModelSchema> {
    const resp = await axios.post<IModelSchema>(`/api/v1/model_repositories/${modelRepositoryName}/models`, data)
    return resp.data
}

export async function startModelUpload(modelRepositoryName: string, version: string): Promise<IModelSchema> {
    const resp = await axios.post<IModelSchema>(
        `/api/v1/model_repositories/${modelRepositoryName}/models/${version}/start_upload`
    )
    return resp.data
}

export async function finishModelUpload(
    modelRepositoryName: string,
    version: string,
    data: IFinishedUploadModelSchema
): Promise<IModelSchema> {
    const resp = await axios.post<IModelSchema>(
        `/api/v1/model_repositories/${modelRepositoryName}/models/${version}/finish_upload`,
        data
    )
    return resp.data
}
