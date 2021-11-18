import axios from 'axios'
import { ICreateApiTokenSchema, IApiTokenSchema, IUpdateApiTokenSchema, IApiTokenFullSchema } from '@/schemas/api_token'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listApiTokens(query: IListQuerySchema): Promise<IListSchema<IApiTokenSchema>> {
    const resp = await axios.get<IListSchema<IApiTokenSchema>>('/api/v1/api_tokens', { params: query })
    return resp.data
}

export async function fetchApiToken(apiTokenUid: string): Promise<IApiTokenFullSchema> {
    const resp = await axios.get<IApiTokenFullSchema>(`/api/v1/api_tokens/${apiTokenUid}`)
    return resp.data
}

export async function createApiToken(data: ICreateApiTokenSchema): Promise<IApiTokenFullSchema> {
    const resp = await axios.post<IApiTokenFullSchema>('/api/v1/api_tokens', data)
    return resp.data
}

export async function updateApiToken(apiTokenUid: string, data: IUpdateApiTokenSchema): Promise<IApiTokenSchema> {
    const resp = await axios.patch<IApiTokenSchema>(`/api/v1/api_tokens/${apiTokenUid}`, data)
    return resp.data
}

export async function deleteApiToken(apiTokenUid: string): Promise<IApiTokenSchema> {
    const resp = await axios.delete<IApiTokenSchema>(`/api/v1/api_tokens/${apiTokenUid}`)
    return resp.data
}
