/* eslint-disable import/no-cycle */
import { IUserSchema } from './user'
import { IBentoFullSchema } from './bento'
import { IResourceSchema } from './resource'

export type DeploymentTargetType = 'stable' | 'canary'
export const DeploymentTargetTypeAddrs: { [k in DeploymentTargetType]: string } = {
    stable: 'stb',
    canary: 'cnr',
}

export interface IDeploymentTargetSchema extends IResourceSchema {
    creator?: IUserSchema
    type: DeploymentTargetType
    bento: IBentoFullSchema
    canary_rules?: IDeploymentTargetCanaryRule[]
    config?: IDeploymentTargetConfigSchema
}

export interface ICreateDeploymentTargetSchema {
    type: DeploymentTargetType
    bento_repository: string
    bento: string
    canary_rules?: IDeploymentTargetCanaryRule[]
    config?: IDeploymentTargetConfigSchema
}

export type DeploymentTargetCanaryRuleType = 'weight' | 'header' | 'cookie'

export interface IDeploymentTargetCanaryRule {
    type: DeploymentTargetCanaryRuleType
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

export interface IDeploymentTargetConfigSchema {
    resources?: IKubeResources
    hpa_conf?: IKubeHPAConf
}
