import { ITerminalRecordSchema } from '@/schemas/terminal_record'
import axios from 'axios'
import { ICreateDeploymentSchema, IDeploymentSchema, IUpdateDeploymentSchema } from '@/schemas/deployment'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listClusterDeployments(
    clusterName: string,
    query: IListQuerySchema
): Promise<IListSchema<IDeploymentSchema>> {
    const resp = await axios.get<IListSchema<IDeploymentSchema>>(`/api/v1/clusters/${clusterName}/deployments`, {
        params: query,
    })
    return resp.data
}

export async function listOrganizationDeployments(query: IListQuerySchema): Promise<IListSchema<IDeploymentSchema>> {
    const resp = await axios.get<IListSchema<IDeploymentSchema>>('/api/v1/deployments', {
        params: query,
    })
    return resp.data
}

export async function listDeploymentTerminalRecords(
    clusterName: string,
    deploymentName: string,
    query: IListQuerySchema
): Promise<IListSchema<ITerminalRecordSchema>> {
    const resp = await axios.get<IListSchema<ITerminalRecordSchema>>(
        `/api/v1/clusters/${clusterName}/deployments/${deploymentName}/terminal_records`,
        {
            params: query,
        }
    )
    return resp.data
}

export async function fetchDeployment(clusterName: string, deploymentName: string): Promise<IDeploymentSchema> {
    const resp = await axios.get<IDeploymentSchema>(`/api/v1/clusters/${clusterName}/deployments/${deploymentName}`)
    return resp.data
}

export async function createDeployment(clusterName: string, data: ICreateDeploymentSchema): Promise<IDeploymentSchema> {
    const resp = await axios.post<IDeploymentSchema>(`/api/v1/clusters/${clusterName}/deployments`, data)
    return resp.data
}

export async function updateDeployment(
    clusterName: string,
    deploymentName: string,
    data: IUpdateDeploymentSchema
): Promise<IDeploymentSchema> {
    const resp = await axios.patch<IDeploymentSchema>(
        `/api/v1/clusters/${clusterName}/deployments/${deploymentName}`,
        data
    )
    return resp.data
}
