import axios from 'axios'
import {
    ICreateOrganizationSchema,
    IOrganizationFullSchema,
    IOrganizationSchema,
    IUpdateOrganizationSchema,
} from '@/schemas/organization'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listOrganizations(query: IListQuerySchema): Promise<IListSchema<IOrganizationSchema>> {
    const resp = await axios.get<IListSchema<IOrganizationSchema>>('/api/v1/orgs', { params: query })
    return resp.data
}

export async function fetchOrganization(orgName: string): Promise<IOrganizationFullSchema> {
    const resp = await axios.get<IOrganizationFullSchema>(`/api/v1/orgs/${orgName}`)
    return resp.data
}

export async function createOrganization(data: ICreateOrganizationSchema): Promise<IOrganizationFullSchema> {
    const resp = await axios.post<IOrganizationFullSchema>('/api/v1/orgs', data)
    return resp.data
}

export async function updateOrganization(
    orgName: string,
    data: IUpdateOrganizationSchema
): Promise<IOrganizationFullSchema> {
    const resp = await axios.patch<IOrganizationFullSchema>(`/api/v1/orgs/${orgName}`, data)
    return resp.data
}
