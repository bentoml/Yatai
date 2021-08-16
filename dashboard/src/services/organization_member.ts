import { ICreateMembersSchema, IDeleteMemberSchema } from '@/schemas/member'
import { IOrganizationMemberSchema } from '@/schemas/organization_member'
import axios from 'axios'

export async function listOrganizationMembers(orgName: string): Promise<IOrganizationMemberSchema[]> {
    const resp = await axios.get<IOrganizationMemberSchema[]>(`/api/v1/orgs/${orgName}/members`)
    return resp.data
}

export async function createOrganizationMembers(
    orgName: string,
    data: ICreateMembersSchema
): Promise<IOrganizationMemberSchema[]> {
    const resp = await axios.post<IOrganizationMemberSchema[]>(`/api/v1/orgs/${orgName}/members`, data)
    return resp.data
}

export async function deleteOrganizationMember(
    orgName: string,
    data: IDeleteMemberSchema
): Promise<IOrganizationMemberSchema> {
    const resp = await axios.delete<IOrganizationMemberSchema>(`/api/v1/orgs/${orgName}/members`, { data })
    return resp.data
}
