import axios from 'axios'
import { ICreateClusterSchema, IClusterSchema, IUpdateClusterSchema, IClusterFullSchema } from '@/schemas/cluster'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listClusters(orgName: string, query: IListQuerySchema): Promise<IListSchema<IClusterSchema>> {
    if (!orgName) {
        return new Promise((resolve) => {
            resolve({
                total: 0,
                start: 0,
                count: 0,
                items: [],
            })
        })
    }
    const resp = await axios.get<IListSchema<IClusterSchema>>(`/api/v1/orgs/${orgName}/clusters`, { params: query })
    return resp.data
}

export async function fetchCluster(orgName: string, clusterName: string): Promise<IClusterFullSchema> {
    const resp = await axios.get<IClusterFullSchema>(`/api/v1/orgs/${orgName}/clusters/${clusterName}`)
    return resp.data
}

export async function createCluster(orgName: string, data: ICreateClusterSchema): Promise<IClusterSchema> {
    const resp = await axios.post<IClusterSchema>(`/api/v1/orgs/${orgName}/clusters`, data)
    return resp.data
}

export async function updateCluster(
    orgName: string,
    clusterName: string,
    data: IUpdateClusterSchema
): Promise<IClusterSchema> {
    const resp = await axios.patch<IClusterSchema>(`/api/v1/orgs/${orgName}/clusters/${clusterName}`, data)
    return resp.data
}
