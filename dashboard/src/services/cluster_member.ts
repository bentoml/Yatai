import { ICreateMembersSchema, IDeleteMemberSchema } from '@/schemas/member'
import { IClusterMemberSchema } from '@/schemas/cluster_member'
import axios from 'axios'

export async function listClusterMembers(orgName: string, clusterName: string): Promise<IClusterMemberSchema[]> {
    const resp = await axios.get<IClusterMemberSchema[]>(`/api/v1/orgs/${orgName}/clusters/${clusterName}/members`)
    return resp.data
}

export async function createClusterMembers(
    orgName: string,
    clusterName: string,
    data: ICreateMembersSchema
): Promise<IClusterMemberSchema[]> {
    const resp = await axios.post<IClusterMemberSchema[]>(
        `/api/v1/orgs/${orgName}/clusters/${clusterName}/members`,
        data
    )
    return resp.data
}

export async function deleteClusterMember(
    orgName: string,
    clusterName: string,
    data: IDeleteMemberSchema
): Promise<IClusterMemberSchema> {
    const resp = await axios.delete<IClusterMemberSchema>(`/api/v1/orgs/${orgName}/clusters/${clusterName}/members`, {
        data,
    })
    return resp.data
}
