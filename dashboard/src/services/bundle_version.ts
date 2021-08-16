import axios from 'axios'
import {
    ICreateBundleVersionSchema,
    IBundleVersionSchema,
    IFinishUploadBundleVersionSchema,
} from '@/schemas/bundle_version'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listBundleVersions(
    orgName: string,
    bundleName: string,
    query: IListQuerySchema
): Promise<IListSchema<IBundleVersionSchema>> {
    const resp = await axios.get<IListSchema<IBundleVersionSchema>>(
        `/api/v1/orgs/${orgName}/bundles/${bundleName}/versions`,
        { params: query }
    )
    return resp.data
}

export async function fetchBundleVersion(
    orgName: string,
    bundleName: string,
    version: string
): Promise<IBundleVersionSchema> {
    const resp = await axios.get<IBundleVersionSchema>(
        `/api/v1/orgs/${orgName}/bundles/${bundleName}/versions/${version}`
    )
    return resp.data
}

export async function createBundleVersion(
    orgName: string,
    bundleName: string,
    data: ICreateBundleVersionSchema
): Promise<IBundleVersionSchema> {
    const resp = await axios.post<IBundleVersionSchema>(`/api/v1/orgs/${orgName}/bundles/${bundleName}/versions`, data)
    return resp.data
}

export async function startBundleVersionUpload(
    orgName: string,
    bundleName: string,
    version: string
): Promise<IBundleVersionSchema> {
    const resp = await axios.post<IBundleVersionSchema>(
        `/api/v1/orgs/${orgName}/bundles/${bundleName}/versions/${version}/start_upload`
    )
    return resp.data
}

export async function finishBundleVersionUpload(
    orgName: string,
    clusterName: string,
    bundleName: string,
    version: string,
    data: IFinishUploadBundleVersionSchema
): Promise<IBundleVersionSchema> {
    const resp = await axios.post<IBundleVersionSchema>(
        `/api/v1/orgs/${orgName}/bundles/${bundleName}/versions/${version}/finish_upload`,
        data
    )
    return resp.data
}
