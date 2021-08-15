import axios from 'axios'
import { IUserSchema, IRegisterUserSchema, ILoginUserSchema } from '@/schemas/user'
import { IListQuerySchema, IListSchema } from '@/schemas/list'

export async function listUsers(query: IListQuerySchema): Promise<IListSchema<IUserSchema>> {
    const resp = await axios.get<IListSchema<IUserSchema>>('/api/v1/users', {
        params: query,
    })
    return resp.data
}

export async function fetchUser(userName: string): Promise<IUserSchema> {
    const resp = await axios.get<IUserSchema>(`/api/v1/users/${userName}`)
    return resp.data
}

export async function fetchCurrentUser(): Promise<IUserSchema> {
    const resp = await axios.get<IUserSchema>('/api/v1/auth/current')
    return resp.data
}

export async function registerUser(data: IRegisterUserSchema): Promise<IUserSchema> {
    const resp = await axios.post<IUserSchema>('/api/v1/auth/register', data)
    return resp.data
}

export async function loginUser(data: ILoginUserSchema): Promise<IUserSchema> {
    const resp = await axios.post<IUserSchema>('/api/v1/auth/login', data)
    return resp.data
}
