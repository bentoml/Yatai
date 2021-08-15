import { IClusterSchema } from './cluster'
import { IResourceSchema } from './resource'
import { IUserSchema } from './user'

export interface IBundleSchema extends IResourceSchema {
    creator?: IUserSchema
    cluster?: IClusterSchema
    description: string
}

export interface ICreateBundleSchema {
    name: string
    description: string
}

export interface IUpdateBundleSchema {
    description?: string
}
