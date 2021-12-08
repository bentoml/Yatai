/* eslint-disable import/no-cycle */
import { IBentoSchema } from './bento'
import { IOrganizationSchema } from './organization'
import { IResourceSchema } from './resource'
import { IUserSchema } from './user'

export interface IBentoRepositorySchema extends IResourceSchema {
    latest_bento?: IBentoSchema
    creator?: IUserSchema
    organization?: IOrganizationSchema
    description: string
}

export interface ICreateBentoRepositorySchema {
    name: string
    description: string
}

export interface IUpdateBentoRepositorySchema {
    description?: string
}
