import { IOrganizationSchema } from './organization'
import { IResourceSchema } from './resource'
import { IUserSchema } from './user'

export type ApiTokenScope = 'api' | 'read_organization' | 'write_organization' | 'read_cluster' | 'write_cluster'

export interface IApiTokenSchema extends IResourceSchema {
    description: string
    user?: IUserSchema
    organization?: IOrganizationSchema
    scopes: ApiTokenScope[]
    expired_at?: string
    last_used_at?: string
    is_expired: boolean
}

export interface IApiTokenFullSchema extends IApiTokenSchema {
    token: string
}

export interface IUpdateApiTokenSchema {
    description?: string
    scopes?: ApiTokenScope[]
    expired_at?: string
}

export interface ICreateApiTokenSchema {
    name: string
    description: string
    scopes: ApiTokenScope[]
    expired_at?: string
}
