/* eslint-disable import/no-cycle */
import { IBentoSchema } from './bento'
import { IDeploymentSchema } from './deployment'
import { IOrganizationSchema } from './organization'
import { IResourceSchema } from './resource'
import { IUserSchema } from './user'

export interface IBentoRepositorySchema extends IResourceSchema {
    latest_bento?: IBentoSchema
    creator?: IUserSchema
    organization?: IOrganizationSchema
    description: string
    n_bentos: number
    n_deployments: number
    latest_bentos: IBentoSchema[]
}

export interface IBentoRepositoryWithLatestDeploymentsSchema extends IBentoRepositorySchema {
    latest_deployments: IDeploymentSchema[]
}

export interface ICreateBentoRepositorySchema {
    name: string
    description: string
}

export interface IUpdateBentoRepositorySchema {
    description?: string
}
