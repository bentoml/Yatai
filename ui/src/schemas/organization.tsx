import { IResourceSchema } from './resource'
import { IUserSchema } from './user'

export interface IOrganizationSchema extends IResourceSchema {
    creator?: IUserSchema
    description: string
}

export interface IUpdateOrganizationSchema {
    description?: string
}

export interface ICreateOrganizationSchema extends IUpdateOrganizationSchema {
    name: string
}
