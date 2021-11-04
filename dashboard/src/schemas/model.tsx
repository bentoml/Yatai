/* eslint-disable import/no-cycle */
import { IModelVersionSchema } from './model_version'
import { IOrganizationSchema } from './organization'
import { IResourceSchema } from './resource'
import { IUserSchema } from './user'

export interface IModelSchema extends IResourceSchema {
    latest_version?: IModelVersionSchema
    creator?: IUserSchema
    organization?: IOrganizationSchema
    description?: string
}

export interface ICreateModelSchema {
    name: string
    description?: string
}

export interface IUpdateModelSchema {
    description?: string
}
