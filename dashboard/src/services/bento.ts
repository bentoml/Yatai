import axios from 'axios'
import {
    ICreateBentoSchema,
    IBentoSchema,
    IFinishUploadBentoSchema,
    IBentoWithRepositorySchema,
    IUpdateBentoSchema,
    IBentoFullSchema,
} from '@/schemas/bento'
import { IListQuerySchema, IListSchema } from '@/schemas/list'
import { IKubePodSchema } from '@/schemas/kube_pod'
import { IModelWithRepositorySchema } from '@/schemas/model'
import { IDeploymentSchema } from '@/schemas/deployment'

export async function listAllBentos(query: IListQuerySchema): Promise<IListSchema<IBentoWithRepositorySchema>> {
    const resp = await axios.get<IListSchema<IBentoWithRepositorySchema>>('/api/v1/bentos', {
        params: query,
    })
    return resp.data
}

export async function listBentos(
    bentoRepositoryName: string,
    query: IListQuerySchema
): Promise<IListSchema<IBentoWithRepositorySchema>> {
    const resp = await axios.get<IListSchema<IBentoWithRepositorySchema>>(
        `/api/v1/bento_repositories/${bentoRepositoryName}/bentos`,
        {
            params: query,
        }
    )
    return resp.data
}

export async function fetchBento(bentoRepositoryName: string, version: string): Promise<IBentoFullSchema> {
    const resp = await axios.get<IBentoFullSchema>(
        `/api/v1/bento_repositories/${bentoRepositoryName}/bentos/${version}`
    )
    return resp.data
}

export async function listBentoModels(
    bentoRepositoryName: string,
    version: string
): Promise<IModelWithRepositorySchema[]> {
    const resp = await axios.get<IModelWithRepositorySchema[]>(
        `/api/v1/bento_repositories/${bentoRepositoryName}/bentos/${version}/models`
    )
    return resp.data
}

export async function listBentoDeployments(
    bentoRepositoryName: string,
    version: string,
    query: IListQuerySchema
): Promise<IListSchema<IDeploymentSchema>> {
    const resp = await axios.get<IListSchema<IDeploymentSchema>>(
        `/api/v1/bento_repositories/${bentoRepositoryName}/bentos/${version}/deployments`,
        {
            params: query,
        }
    )
    return resp.data
}

export async function createBento(bentoRepositoryName: string, data: ICreateBentoSchema): Promise<IBentoSchema> {
    const resp = await axios.post<IBentoSchema>(`/api/v1/bento_repositories/${bentoRepositoryName}/bentos`, data)
    return resp.data
}

export async function updateBento(
    bentoRepositoryName: string,
    version: string,
    data: IUpdateBentoSchema
): Promise<IBentoFullSchema> {
    const response = await axios.patch<IBentoFullSchema>(
        `/api/v1/bento_repositories/${bentoRepositoryName}/bentos/${version}`,
        data
    )
    return response.data
}

export async function startBentoUpload(bentoRepositoryName: string, version: string): Promise<IBentoSchema> {
    const resp = await axios.post<IBentoSchema>(
        `/api/v1/bento_repositories/${bentoRepositoryName}/bentos/${version}/start_upload`
    )
    return resp.data
}

export async function finishBentoUpload(
    bentoRepositoryName: string,
    version: string,
    data: IFinishUploadBentoSchema
): Promise<IBentoSchema> {
    const resp = await axios.post<IBentoSchema>(
        `/api/v1/bento_repositories/${bentoRepositoryName}/bentos/${version}/finish_upload`,
        data
    )
    return resp.data
}

export async function recreateBentoImageBuilderJob(
    bentoRepositoryName: string,
    version: string
): Promise<IBentoSchema> {
    const resp = await axios.patch<IBentoSchema>(
        `/api/v1/bento_repositories/${bentoRepositoryName}/bentos/${version}/recreate_image_builder_job`
    )
    return resp.data
}

export async function listBentoImageBuilderPods(
    bentoRepositoryName: string,
    version: string
): Promise<IKubePodSchema[]> {
    const resp = await axios.post<IKubePodSchema[]>(
        `/api/v1/bento_repositories/${bentoRepositoryName}/bentos/${version}/image_builder_pods`
    )
    return resp.data
}
