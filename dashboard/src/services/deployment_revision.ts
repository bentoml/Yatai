import axios from 'axios'
import { IDeploymentRevisionSchema } from '@/schemas/deployment_revision'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listDeploymentRevisions(
    clusterName: string,
    deploymentName: string,
    query: IListQuerySchema
): Promise<IListSchema<IDeploymentRevisionSchema>> {
    const resp = await axios.get<IListSchema<IDeploymentRevisionSchema>>(
        `/api/v1/clusters/${clusterName}/deployments/${deploymentName}/revisions`,
        {
            params: query,
        }
    )
    return resp.data
}
