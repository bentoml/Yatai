import axios from 'axios'
import {
    ICreateBentoVersionSchema,
    IBentoVersionSchema,
    IFinishUploadBentoVersionSchema,
    IBentoVersionWithBentoSchema,
} from '@/schemas/bento_version'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listAllBentoVersions(
    query: IListQuerySchema
): Promise<IListSchema<IBentoVersionWithBentoSchema>> {
    const resp = await axios.get<IListSchema<IBentoVersionWithBentoSchema>>('/api/v1/bento_versions', {
        params: query,
    })
    return resp.data
}

export async function listBentoVersions(
    bentoName: string,
    query: IListQuerySchema
): Promise<IListSchema<IBentoVersionSchema>> {
    const resp = await axios.get<IListSchema<IBentoVersionSchema>>(`/api/v1/bentos/${bentoName}/versions`, {
        params: query,
    })
    return resp.data
}

export async function fetchBentoVersion(bentoName: string, version: string): Promise<IBentoVersionSchema> {
    const resp = await axios.get<IBentoVersionSchema>(`/api/v1/bentos/${bentoName}/versions/${version}`)
    return resp.data
}

export async function createBentoVersion(
    bentoName: string,
    data: ICreateBentoVersionSchema
): Promise<IBentoVersionSchema> {
    const resp = await axios.post<IBentoVersionSchema>(`/api/v1/bentos/${bentoName}/versions`, data)
    return resp.data
}

export async function startBentoVersionUpload(bentoName: string, version: string): Promise<IBentoVersionSchema> {
    const resp = await axios.post<IBentoVersionSchema>(`/api/v1/bentos/${bentoName}/versions/${version}/start_upload`)
    return resp.data
}

export async function finishBentoVersionUpload(
    bentoName: string,
    version: string,
    data: IFinishUploadBentoVersionSchema
): Promise<IBentoVersionSchema> {
    const resp = await axios.post<IBentoVersionSchema>(
        `/api/v1/bentos/${bentoName}/versions/${version}/finish_upload`,
        data
    )
    return resp.data
}
