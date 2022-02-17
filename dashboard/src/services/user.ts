import axios from 'axios'
import { IUserSchema, IRegisterUserSchema, ILoginUserSchema, ICreateUserSchema } from '@/schemas/user'
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
    console.log('we are fetching the current user') // eslint-disable-line no-console
    return resp.data
}

export async function registerUser(data: IRegisterUserSchema): Promise<IUserSchema> {
    const resp = await axios.post<IUserSchema>('/api/v1/auth/register', data)
    console.log(resp.data) // eslint-disable-line
    return resp.data
}

export async function loginUser(data: ILoginUserSchema): Promise<IUserSchema> {
    const resp = await axios.post<IUserSchema>('/api/v1/auth/login', data)
    return resp.data
}

export async function fetchCurrentUserApiToken(): Promise<IUserSchema> {
    const resp = await axios.put<IUserSchema>('/api/v1/auth/current/api_token')
    return resp.data
}

export async function createUser(data: ICreateUserSchema): Promise<IUserSchema> {
    const resp = await axios.post<IUserSchema>('/api/v1/users', data)
    return resp.data
}
