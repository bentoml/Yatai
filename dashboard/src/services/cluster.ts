import axios from 'axios'
import { ICreateClusterSchema, IClusterSchema, IUpdateClusterSchema, IClusterFullSchema } from '@/schemas/cluster'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listClusters(query: IListQuerySchema): Promise<IListSchema<IClusterSchema>> {
    const resp = await axios.get<IListSchema<IClusterSchema>>('/api/v1/clusters', { params: query })
    return resp.data
}

export async function fetchCluster(clusterName: string): Promise<IClusterFullSchema> {
    const resp = await axios.get<IClusterFullSchema>(`/api/v1/clusters/${clusterName}`)
    return resp.data
}

export async function createCluster(data: ICreateClusterSchema): Promise<IClusterFullSchema> {
    const resp = await axios.post<IClusterFullSchema>('/api/v1/clusters', data)
    return resp.data
}

export async function updateCluster(clusterName: string, data: IUpdateClusterSchema): Promise<IClusterFullSchema> {
    const resp = await axios.patch<IClusterFullSchema>(`/api/v1/clusters/${clusterName}`, data)
    return resp.data
}
