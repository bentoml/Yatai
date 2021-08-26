import axios from 'axios'
import { ICreateBentoSchema, IBentoSchema, IUpdateBentoSchema } from '@/schemas/bento'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listBentos(orgName: string, query: IListQuerySchema): Promise<IListSchema<IBentoSchema>> {
    const resp = await axios.get<IListSchema<IBentoSchema>>(`/api/v1/orgs/${orgName}/bentos`, { params: query })
    return resp.data
}

export async function fetchBento(orgName: string, bentoName: string): Promise<IBentoSchema> {
    const resp = await axios.get<IBentoSchema>(`/api/v1/orgs/${orgName}/bentos/${bentoName}`)
    return resp.data
}

export async function createBento(orgName: string, data: ICreateBentoSchema): Promise<IBentoSchema> {
    const resp = await axios.post<IBentoSchema>(`/api/v1/orgs/${orgName}/bentos`, data)
    return resp.data
}

export async function updateBento(orgName: string, bentoName: string, data: IUpdateBentoSchema): Promise<IBentoSchema> {
    const resp = await axios.patch<IBentoSchema>(`/api/v1/orgs/${orgName}/bentos/${bentoName}`, data)
    return resp.data
}
