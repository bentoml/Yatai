import axios from 'axios'
import {
    ICreateBentoVersionSchema,
    IBentoVersionSchema,
    IFinishUploadBentoVersionSchema,
} from '@/schemas/bento_version'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listBentoVersions(
    orgName: string,
    bentoName: string,
    query: IListQuerySchema
): Promise<IListSchema<IBentoVersionSchema>> {
    const resp = await axios.get<IListSchema<IBentoVersionSchema>>(
        `/api/v1/orgs/${orgName}/bentos/${bentoName}/versions`,
        { params: query }
    )
    return resp.data
}

export async function fetchBentoVersion(
    orgName: string,
    bentoName: string,
    version: string
): Promise<IBentoVersionSchema> {
    const resp = await axios.get<IBentoVersionSchema>(`/api/v1/orgs/${orgName}/bentos/${bentoName}/versions/${version}`)
    return resp.data
}

export async function createBentoVersion(
    orgName: string,
    bentoName: string,
    data: ICreateBentoVersionSchema
): Promise<IBentoVersionSchema> {
    const resp = await axios.post<IBentoVersionSchema>(`/api/v1/orgs/${orgName}/bentos/${bentoName}/versions`, data)
    return resp.data
}

export async function startBentoVersionUpload(
    orgName: string,
    bentoName: string,
    version: string
): Promise<IBentoVersionSchema> {
    const resp = await axios.post<IBentoVersionSchema>(
        `/api/v1/orgs/${orgName}/bentos/${bentoName}/versions/${version}/start_upload`
    )
    return resp.data
}

export async function finishBentoVersionUpload(
    orgName: string,
    clusterName: string,
    bentoName: string,
    version: string,
    data: IFinishUploadBentoVersionSchema
): Promise<IBentoVersionSchema> {
    const resp = await axios.post<IBentoVersionSchema>(
        `/api/v1/orgs/${orgName}/bentos/${bentoName}/versions/${version}/finish_upload`,
        data
    )
    return resp.data
}
