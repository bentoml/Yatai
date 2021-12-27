import { IDeploymentSchema } from '@/schemas/deployment'
import { IBentoWithRepositorySchema } from '@/schemas/bento'
import axios from 'axios'

import {
    IModelSchema,
    IModelFullSchema,
    ICreateModelSchema,
    IFinishedUploadModelSchema,
    IModelWithRepositorySchema,
    IUpdateModelSchema,
} from '@/schemas/model'
import { IListQuerySchema, IListSchema } from '@/schemas/list'
import { IKubePodSchema } from '@/schemas/kube_pod'

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

export async function fetchModel(modelRepositoryName: string, version: string): Promise<IModelFullSchema> {
    const resp = await axios.get<IModelFullSchema>(
        `/api/v1/model_repositories/${modelRepositoryName}/models/${version}`
    )
    return resp.data
}

export async function listModelBentos(
    modelRepositoryName: string,
    version: string,
    query: IListQuerySchema
): Promise<IListSchema<IBentoWithRepositorySchema>> {
    const resp = await axios.get<IListSchema<IBentoWithRepositorySchema>>(
        `/api/v1/model_repositories/${modelRepositoryName}/models/${version}/bentos`,
        {
            params: query,
        }
    )
    return resp.data
}

export async function listModelDeployments(
    modelRepositoryName: string,
    version: string,
    query: IListQuerySchema
): Promise<IListSchema<IDeploymentSchema>> {
    const resp = await axios.get<IListSchema<IDeploymentSchema>>(
        `/api/v1/model_repositories/${modelRepositoryName}/models/${version}/deployments`,
        {
            params: query,
        }
    )
    return resp.data
}

export async function createModel(modelRepositoryName: string, data: ICreateModelSchema): Promise<IModelSchema> {
    const resp = await axios.post<IModelSchema>(`/api/v1/model_repositories/${modelRepositoryName}/models`, data)
    return resp.data
}

export async function updateModel(
    modelRepositoryName: string,
    version: string,
    data: IUpdateModelSchema
): Promise<IModelFullSchema> {
    const response = await axios.patch<IModelFullSchema>(
        `/api/v1/model_repositories/${modelRepositoryName}/models/${version}`,
        data
    )
    return response.data
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

export async function recreateModelImageBuilderJob(
    modelRepositoryName: string,
    version: string
): Promise<IModelSchema> {
    const resp = await axios.patch<IModelSchema>(
        `/api/v1/model_repositories/${modelRepositoryName}/models/${version}/recreate_image_builder_job`
    )
    return resp.data
}

export async function listModelImageBuilderPods(
    modelRepositoryName: string,
    version: string
): Promise<IKubePodSchema[]> {
    const resp = await axios.post<IKubePodSchema[]>(
        `/api/v1/model_repositories/${modelRepositoryName}/models/${version}/image_builder_pods`
    )
    return resp.data
}
