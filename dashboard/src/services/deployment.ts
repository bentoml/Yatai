import axios from 'axios'
import { ICreateDeploymentSchema, IDeploymentSchema, IUpdateDeploymentSchema } from '@/schemas/deployment'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listDeployments(
    orgName: string,
    clusterName: string,
    query: IListQuerySchema
): Promise<IListSchema<IDeploymentSchema>> {
    const resp = await axios.get<IListSchema<IDeploymentSchema>>(
        `/api/v1/orgs/${orgName}/clusters/${clusterName}/deployments`,
        {
            params: query,
        }
    )
    return resp.data
}

export async function fetchDeployment(
    orgName: string,
    clusterName: string,
    deploymentName: string
): Promise<IDeploymentSchema> {
    const resp = await axios.get<IDeploymentSchema>(
        `/api/v1/orgs/${orgName}/clusters/${clusterName}/deployments/${deploymentName}`
    )
    return resp.data
}

export async function createDeployment(
    orgName: string,
    clusterName: string,
    data: ICreateDeploymentSchema
): Promise<IDeploymentSchema> {
    const resp = await axios.post<IDeploymentSchema>(
        `/api/v1/orgs/${orgName}/clusters/${clusterName}/deployments`,
        data
    )
    return resp.data
}

export async function updateDeployment(
    orgName: string,
    clusterName: string,
    deploymentName: string,
    data: IUpdateDeploymentSchema
): Promise<IDeploymentSchema> {
    const resp = await axios.patch<IDeploymentSchema>(
        `/api/v1/orgs/${orgName}/clusters/${clusterName}/deployments/${deploymentName}`,
        data
    )
    return resp.data
}
