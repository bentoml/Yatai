import { IOrganizationSchema } from './organization'
import { IResourceSchema } from './resource'
import { IUserSchema } from './user'

export interface IClusterSchema extends IResourceSchema {
    creator?: IUserSchema
    description: string
}

export interface IInfraMinIOSchema {
    endpoint: string
    access_key: string
    secret_key: string
}

export interface IClusterConfigSchema {
    ingress_ip: string
    infra_minio?: IInfraMinIOSchema
}

export interface IClusterFullSchema extends IClusterSchema {
    organization?: IOrganizationSchema
    kube_config: string
    config: IClusterConfigSchema
}

export interface IUpdateClusterSchema {
    description?: string
    kube_config?: string
    config?: IClusterConfigSchema
}

export interface ICreateClusterSchema {
    name: string
    description: string
    kube_config: string
    config: IClusterConfigSchema
}
