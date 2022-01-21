import { IOrganizationSchema } from './organization'
import { IResourceSchema } from './resource'
import { IUserSchema } from './user'

export interface IClusterSchema extends IResourceSchema {
    creator?: IUserSchema
    description: string
}

export interface IClusterConfigSchema {
    ingress_ip: string
    default_deployment_kube_namespace: string
}

export interface IClusterFullSchema extends IClusterSchema {
    organization?: IOrganizationSchema
    kube_config: string
    config: IClusterConfigSchema
    grafana_root_path: string
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
