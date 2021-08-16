import axios from 'axios'
import { ICreateBundleSchema, IBundleSchema, IUpdateBundleSchema } from '@/schemas/bundle'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listBundles(orgName: string, query: IListQuerySchema): Promise<IListSchema<IBundleSchema>> {
    const resp = await axios.get<IListSchema<IBundleSchema>>(`/api/v1/orgs/${orgName}/bundles`, { params: query })
    return resp.data
}

export async function fetchBundle(orgName: string, bundleName: string): Promise<IBundleSchema> {
    const resp = await axios.get<IBundleSchema>(`/api/v1/orgs/${orgName}/bundles/${bundleName}`)
    return resp.data
}

export async function createBundle(orgName: string, data: ICreateBundleSchema): Promise<IBundleSchema> {
    const resp = await axios.post<IBundleSchema>(`/api/v1/orgs/${orgName}/bundles`, data)
    return resp.data
}

export async function updateBundle(
    orgName: string,
    bundleName: string,
    data: IUpdateBundleSchema
): Promise<IBundleSchema> {
    const resp = await axios.patch<IBundleSchema>(`/api/v1/orgs/${orgName}/bundles/${bundleName}`, data)
    return resp.data
}
