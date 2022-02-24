import { MemberRole } from './member_role'
import { IResourceSchema } from './resource'

export interface IUserSchema extends IResourceSchema {
    first_name: string
    last_name: string
    email: string
    avatar_url: string
    api_token: string
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

export interface ICreateUserSchema {
    name: string
    email: string
    password: string
    role: MemberRole
}
export interface IChangePasswordSchema {
    current_password: string
    new_password: string
}
