import { IResourceSchema } from './resource'

export interface IUserSchema extends IResourceSchema {
    first_name: string
    last_name: string
    email: string
    avatar_url: string
}

export interface IRegisterUserSchema {
    name: string
    first_name: string
    last_name: string
    email: string
    password: string
}

export interface ILoginUserSchema {
    name_or_email: string
    password: string
}

export interface IUpdateUserSchema {
    first_name: string
    last_name: string
}
