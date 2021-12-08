/* eslint-disable import/no-cycle */
import { IModelSchema } from './model'
import { IOrganizationSchema } from './organization'
import { IResourceSchema } from './resource'
import { IUserSchema } from './user'

export interface IModelRepositorySchema extends IResourceSchema {
    latest_model?: IModelSchema
    creator?: IUserSchema
    organization?: IOrganizationSchema
    description?: string
}

export interface ICreateModelRepositorySchema {
    name: string
    description?: string
}

export interface IUpdateModelRepositorySchema {
    description?: string
}
