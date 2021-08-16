import { IResourceSchema } from './resource'
import { IUserSchema } from './user'

export interface IInfraMinIOSchema {
    endpoint: string
    access_key: string
    secret_key: string
}

export interface IOrganizationConfigSchema {
    infra_minio?: IInfraMinIOSchema
}

export interface IOrganizationSchema extends IResourceSchema {
    creator?: IUserSchema
    description: string
}

export interface IOrganizationFullSchema extends IOrganizationSchema {
    config?: IOrganizationConfigSchema
}

export interface IUpdateOrganizationSchema {
    description?: string
}

export interface ICreateOrganizationSchema extends IUpdateOrganizationSchema {
    name: string
}
