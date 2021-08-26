import { IResourceSchema } from './resource'
import { IUserSchema } from './user'

export interface IAwsS3Schema {
    bucket_name: string
    region: string
}

export interface IAwsECRSchema {
    repository_uri: string
    region: string
}

export interface IOrganizationAwsConfigSchema {
    access_key_id: string
    secret_access_key: string
    s3?: IAwsS3Schema
    ecr?: IAwsECRSchema
}

export interface IOrganizationConfigSchema {
    major_cluster_uid?: string
    aws?: IOrganizationAwsConfigSchema
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
    config?: IOrganizationConfigSchema
}

export interface ICreateOrganizationSchema {
    name: string
    description: string
    config?: IOrganizationConfigSchema
}
