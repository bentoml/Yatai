import axios from 'axios'
import { ICreateBentoSchema, IBentoSchema, IUpdateBentoSchema } from '@/schemas/bento'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listBentos(query: IListQuerySchema): Promise<IListSchema<IBentoSchema>> {
    const resp = await axios.get<IListSchema<IBentoSchema>>('/api/v1/bentos', { params: query })
    return resp.data
}

export async function fetchBento(bentoName: string): Promise<IBentoSchema> {
    const resp = await axios.get<IBentoSchema>(`/api/v1/bentos/${bentoName}`)
    return resp.data
}

export async function createBento(data: ICreateBentoSchema): Promise<IBentoSchema> {
    const resp = await axios.post<IBentoSchema>('/api/v1/bentos', data)
    return resp.data
}

export async function updateBento(bentoName: string, data: IUpdateBentoSchema): Promise<IBentoSchema> {
    const resp = await axios.patch<IBentoSchema>(`/api/v1/bentos/${bentoName}`, data)
    return resp.data
}
