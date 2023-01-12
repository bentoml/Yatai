/* eslint-disable import/no-cycle */
import { ICreateDeploymentTargetSchema, ICreateDeploymentTargetUISchema } from './deployment_target'
import { IDeploymentRevisionSchema } from './deployment_revision'
import { IUserSchema } from './user'
import { IResourceSchema } from './resource'
import { IClusterFullSchema } from './cluster'

export type DeploymentStatus =
    | 'unknown'
    | 'non-deployed'
    | 'running'
    | 'unhealthy'
    | 'failed'
    | 'deploying'
    | 'terminating'
    | 'terminated'
    | 'image-building'
    | 'image-build-failed'
    | 'image-build-succeeded'

export interface IDeploymentSchema extends IResourceSchema {
    description: string
    creator?: IUserSchema
    cluster?: IClusterFullSchema
    status: DeploymentStatus
    urls: string[]
    latest_revision?: IDeploymentRevisionSchema
    kube_namespace: string
}

// eslint-disable-next-line @typescript-eslint/no-empty-interface
export interface IDeploymentFullSchema extends IDeploymentSchema {}

export interface IUpdateDeploymentUISchema {
    targets: ICreateDeploymentTargetUISchema[]
}

export interface IUpdateDeploymentSchema {
    targets: ICreateDeploymentTargetSchema[]
}

export interface ICreateDeploymentUISchema extends IUpdateDeploymentUISchema {
    cluster_name?: string
    name: string
    description: string
    kube_namespace?: string
}

export interface ICreateDeploymentSchema extends IUpdateDeploymentSchema {
    cluster_name?: string
    name: string
    description: string
    kube_namespace?: string
}
