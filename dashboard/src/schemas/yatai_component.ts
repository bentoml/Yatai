/* eslint-disable import/no-cycle */
import { IUserSchema } from './user'
import { IResourceSchema } from './resource'
import { IClusterFullSchema } from './cluster'

export interface IYataiComponentSchema extends IResourceSchema {
    description: string
    creator?: IUserSchema
    cluster?: IClusterFullSchema
    kube_namespace: string
    lastest_installed_at?: string
    latest_heartbeat_at?: string
}
