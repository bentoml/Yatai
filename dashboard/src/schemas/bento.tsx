import { IBentoVersionSchema } from './bento_version'
import { IOrganizationSchema } from './organization'
import { IResourceSchema } from './resource'
import { IUserSchema } from './user'

export interface IBentoSchema extends IResourceSchema {
    latest_version?: IBentoVersionSchema
    creator?: IUserSchema
    organization?: IOrganizationSchema
    description: string
}

export interface ICreateBentoSchema {
    name: string
    description: string
}

export interface IUpdateBentoSchema {
    description?: string
}
