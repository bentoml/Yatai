import axios from 'axios'
import { ICreateOrganizationSchema, IOrganizationSchema, IUpdateOrganizationSchema } from '@/schemas/organization'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listOrganizations(query: IListQuerySchema): Promise<IListSchema<IOrganizationSchema>> {
    const resp = await axios.get<IListSchema<IOrganizationSchema>>('/api/v1/orgs', { params: query })
    return resp.data
}

export async function fetchOrganization(orgName: string): Promise<IOrganizationSchema> {
    const resp = await axios.get<IOrganizationSchema>(`/api/v1/orgs/${orgName}`)
    return resp.data
}

export async function createOrganization(data: ICreateOrganizationSchema): Promise<IOrganizationSchema> {
    const resp = await axios.post<IOrganizationSchema>('/api/v1/orgs', data)
    return resp.data
}

export async function updateOrganization(
    orgName: string,
    data: IUpdateOrganizationSchema
): Promise<IOrganizationSchema> {
    const resp = await axios.patch<IOrganizationSchema>(`/api/v1/orgs/${orgName}`, data)
    return resp.data
}
