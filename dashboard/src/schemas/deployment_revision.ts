/* eslint-disable import/no-cycle */
import { IDeploymentTargetSchema } from './deployment_target'
import { IUserSchema } from './user'
import { IResourceSchema } from './resource'

export type DeploymentRevisionStatus = 'active' | 'inactive'

export interface IDeploymentRevisionSchema extends IResourceSchema {
    creator?: IUserSchema
    status: DeploymentRevisionStatus
    targets: IDeploymentTargetSchema[]
}
