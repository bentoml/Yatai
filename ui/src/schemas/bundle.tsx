import { IOrganizationSchema } from './organization'
import { IResourceSchema } from './resource'
import { IUserSchema } from './user'

export interface IBundleSchema extends IResourceSchema {
    creator?: IUserSchema
    organization?: IOrganizationSchema
    description: string
}

export interface ICreateBundleSchema {
    name: string
    description: string
}

export interface IUpdateBundleSchema {
    description?: string
}
