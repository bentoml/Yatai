import { ICreateMembersSchema, IDeleteMemberSchema } from '@/schemas/member'
import { IOrganizationMemberSchema } from '@/schemas/organization_member'
import axios from 'axios'

export async function listOrganizationMembers(): Promise<IOrganizationMemberSchema[]> {
    const resp = await axios.get<IOrganizationMemberSchema[]>('/api/v1/members')
    return resp.data
}

export async function createOrganizationMembers(data: ICreateMembersSchema): Promise<IOrganizationMemberSchema[]> {
    const resp = await axios.post<IOrganizationMemberSchema[]>('/api/v1/members', data)
    return resp.data
}

export async function deleteOrganizationMember(data: IDeleteMemberSchema): Promise<IOrganizationMemberSchema> {
    const resp = await axios.delete<IOrganizationMemberSchema>('/api/v1/members', { data })
    return resp.data
}
