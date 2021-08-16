import { IOrganizationSchema } from './organization'
import { IResourceSchema } from './resource'
import { IUserSchema } from './user'

export interface IClusterSchema extends IResourceSchema {
    creator?: IUserSchema
    description: string
}

export interface IClusterConfigSchema {
    ingress_ip: string
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
