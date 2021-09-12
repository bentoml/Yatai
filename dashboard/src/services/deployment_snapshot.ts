import axios from 'axios'
import { IDeploymentSnapshotSchema } from '@/schemas/deployment_snapshot'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listDeploymentSnapshots(
    orgName: string,
    clusterName: string,
    deploymentName: string,
    query: IListQuerySchema
): Promise<IListSchema<IDeploymentSnapshotSchema>> {
    const resp = await axios.get<IListSchema<IDeploymentSnapshotSchema>>(
        `/api/v1/orgs/${orgName}/clusters/${clusterName}/deployments/${deploymentName}/snapshots`,
        {
            params: query,
        }
    )
    return resp.data
}
