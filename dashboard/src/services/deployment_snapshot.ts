import axios from 'axios'
import { IDeploymentSnapshotSchema } from '@/schemas/deployment_snapshot'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listDeploymentSnapshots(
    clusterName: string,
    deploymentName: string,
    query: IListQuerySchema
): Promise<IListSchema<IDeploymentSnapshotSchema>> {
    const resp = await axios.get<IListSchema<IDeploymentSnapshotSchema>>(
        `/api/v1/clusters/${clusterName}/deployments/${deploymentName}/snapshots`,
        {
            params: query,
        }
    )
    return resp.data
}
