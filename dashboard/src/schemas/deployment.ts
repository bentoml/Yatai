import { ICreateDeploymentTargetSchema } from './deployment_target'
import { IDeploymentRevisionSchema } from './deployment_revision'
import { IUserSchema } from './user'
import { IResourceSchema } from './resource'
import { IClusterFullSchema } from './cluster'

export type DeploymentStatus = 'unknown' | 'non-deployed' | 'running' | 'unhealthy' | 'failed' | 'deploying'

export interface IDeploymentSchema extends IResourceSchema {
    description: string
    creator?: IUserSchema
    cluster?: IClusterFullSchema
    status: DeploymentStatus
    urls: string[]
    latest_revision?: IDeploymentRevisionSchema
}

// eslint-disable-next-line @typescript-eslint/no-empty-interface
export interface IDeploymentFullSchema extends IDeploymentSchema {}

export interface IUpdateDeploymentSchema {
    targets: ICreateDeploymentTargetSchema[]
}

export interface ICreateDeploymentSchema extends IUpdateDeploymentSchema {
    cluster_name?: string
    name: string
    description: string
}
