import {
    DeploymentSnapshotType,
    IDeploymentSnapshotCanaryRule,
    IDeploymentSnapshotConfigSchema,
    IDeploymentSnapshotSchema,
} from './deployment_snapshot'
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
}

export interface IDeploymentFullSchema extends IDeploymentSchema {
    latest_snapshot?: IDeploymentSnapshotSchema
}

export interface IUpdateDeploymentSchema {
    type: DeploymentSnapshotType
    bento_name: string
    bento_version: string
    canary_rules?: IDeploymentSnapshotCanaryRule[]
    config?: IDeploymentSnapshotConfigSchema
}

export interface ICreateDeploymentSchema extends IUpdateDeploymentSchema {
    cluster_name?: string
    name: string
    description: string
}
