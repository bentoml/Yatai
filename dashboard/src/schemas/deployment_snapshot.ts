import { IUserSchema } from './user'
import { IBentoVersionFullSchema } from './bento_version'
import { IResourceSchema } from './resource'

export type DeploymentSnapshotStatus = 'active' | 'inactive'
export type DeploymentSnapshotType = 'stable' | 'canary'
export const DeploymentSnapshotTypeAddrs: { [k in DeploymentSnapshotType]: string } = {
    stable: 'stb',
    canary: 'cnr',
}

export interface IDeploymentSnapshotSchema extends IResourceSchema {
    creator?: IUserSchema
    type: DeploymentSnapshotType
    status: DeploymentSnapshotStatus
    bento_version: IBentoVersionFullSchema
    canary_rules?: IDeploymentSnapshotCanaryRule[]
    config?: IDeploymentSnapshotConfigSchema
}

export type DeploymentSnapshotCanaryRuleType = 'weight' | 'header' | 'cookie'

export interface IDeploymentSnapshotCanaryRule {
    type: DeploymentSnapshotCanaryRuleType
    weight?: number
    header?: string
    cookie?: string
    header_value?: string
}

export interface IKubeResourceItem {
    cpu: string
    memory: string
    gpu: string
}

export interface IKubeResources {
    requests?: IKubeResourceItem
    limits?: IKubeResourceItem
}

export interface IRollingUpgradeStrategy {
    max_surge?: string
    max_unavailable?: string
}

export interface IKubeHPAConf {
    cpu?: number
    gpu?: number
    memory?: string
    qps?: number
    max_replicas?: number
    min_replicas?: number
}

export interface IDeploymentSnapshotConfigSchema {
    resources?: IKubeResources
    hpa_conf?: IKubeHPAConf
}
