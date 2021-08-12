import axios from 'axios'
import { ICreateBundleSchema, IBundleSchema, IUpdateBundleSchema } from '@/schemas/bundle'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listBundles(
    orgName: string,
    clusterName: string,
    query: IListQuerySchema
): Promise<IListSchema<IBundleSchema>> {
    const resp = await axios.get<IListSchema<IBundleSchema>>(
        `/api/v1/orgs/${orgName}/clusters/${clusterName}/bundles`,
        { params: query }
    )
    return resp.data
}

export async function fetchBundle(orgName: string, clusterName: string, bundleName: string): Promise<IBundleSchema> {
    const resp = await axios.get<IBundleSchema>(`/api/v1/orgs/${orgName}/clusters/${clusterName}/bundles/${bundleName}`)
    return resp.data
}

export async function createBundle(
    orgName: string,
    clusterName: string,
    data: ICreateBundleSchema
): Promise<IBundleSchema> {
    const resp = await axios.post<IBundleSchema>(`/api/v1/orgs/${orgName}/clusters/${clusterName}/bundles`, data)
    return resp.data
}

export async function updateBundle(
    orgName: string,
    clusterName: string,
    bundleName: string,
    data: IUpdateBundleSchema
): Promise<IBundleSchema> {
    const resp = await axios.patch<IBundleSchema>(
        `/api/v1/orgs/${orgName}/clusters/${clusterName}/bundles/${bundleName}`,
        data
    )
    return resp.data
}
