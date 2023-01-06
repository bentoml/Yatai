/* eslint-disable import/no-cycle */
import { IUserSchema } from './user'
import { IResourceSchema } from './resource'
import { IClusterFullSchema } from './cluster'

export type YataiComponentType = 'deployment' | 'image-builder'

export interface IYataiComponentSchema extends IResourceSchema {
    description: string
    creator?: IUserSchema
    cluster?: IClusterFullSchema
    kube_namespace: string
    latest_installed_at?: string
    latest_heartbeat_at?: string
    version: string
    manifest?: {
        selector_labels: Record<string, string>
        latest_crd_version?: string
    }
}
