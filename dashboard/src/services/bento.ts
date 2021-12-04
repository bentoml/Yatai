import axios from 'axios'
import { ICreateBentoSchema, IBentoSchema, IFinishUploadBentoSchema, IBentoWithRepositorySchema } from '@/schemas/bento'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listAllBentos(query: IListQuerySchema): Promise<IListSchema<IBentoWithRepositorySchema>> {
    const resp = await axios.get<IListSchema<IBentoWithRepositorySchema>>('/api/v1/bentos', {
        params: query,
    })
    return resp.data
}

export async function listBentos(
    bentoRepositoryName: string,
    query: IListQuerySchema
): Promise<IListSchema<IBentoSchema>> {
    const resp = await axios.get<IListSchema<IBentoSchema>>(
        `/api/v1/bento_repositories/${bentoRepositoryName}/bentos`,
        {
            params: query,
        }
    )
    return resp.data
}

export async function fetchBento(bentoRepositoryName: string, version: string): Promise<IBentoSchema> {
    const resp = await axios.get<IBentoSchema>(`/api/v1/bento_repositories/${bentoRepositoryName}/bentos/${version}`)
    return resp.data
}

export async function createBento(bentoRepositoryName: string, data: ICreateBentoSchema): Promise<IBentoSchema> {
    const resp = await axios.post<IBentoSchema>(`/api/v1/bento_repositories/${bentoRepositoryName}/bentos`, data)
    return resp.data
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
