import { IClusterFullSchema } from '@/schemas/cluster'
import axios from 'axios'
import {
    ICreateOrganizationSchema,
    IOrganizationFullSchema,
    IOrganizationSchema,
    IUpdateOrganizationSchema,
} from '@/schemas/organization'
import { IListQuerySchema, IListSchema } from '@/schemas/list'
import { IEventSchema } from '@/schemas/event'
import { ResourceType } from '@/schemas/resource'

export async function listOrganizations(query: IListQuerySchema): Promise<IListSchema<IOrganizationSchema>> {
    const resp = await axios.get<IListSchema<IOrganizationSchema>>('/api/v1/orgs', { params: query })
    return resp.data
}

export async function fetchOrganization(): Promise<IOrganizationFullSchema> {
    const resp = await axios.get<IOrganizationFullSchema>('/api/v1/current_org')
    return resp.data
}

export async function listOrganizationModelModules(): Promise<string[]> {
    const resp = await axios.get<string[]>('/api/v1/current_org/model_modules')
    return resp.data
}

export async function fetchOrganizationMajorCluster(): Promise<IClusterFullSchema> {
    const resp = await axios.get<IClusterFullSchema>('/api/v1/current_org/major_cluster')
    return resp.data
}

export async function listOrganizationEvents(query: IListQuerySchema): Promise<IListSchema<IEventSchema>> {
    const resp = await axios.get<IListSchema<IEventSchema>>('/api/v1/current_org/events', { params: query })
    return resp.data
}

export async function listOrganizationEventOperationNames(resourceType: ResourceType): Promise<string[]> {
    const resp = await axios.get<string[]>('/api/v1/current_org/event_operation_names', {
        params: {
            resource_type: resourceType,
        },
    })
    return resp.data
}

export async function createOrganization(data: ICreateOrganizationSchema): Promise<IOrganizationFullSchema> {
    const resp = await axios.post<IOrganizationFullSchema>('/api/v1/orgs', data)
    return resp.data
}

export async function updateOrganization(data: IUpdateOrganizationSchema): Promise<IOrganizationFullSchema> {
    const resp = await axios.patch<IOrganizationFullSchema>('/api/v1/current_org', data)
    return resp.data
}
