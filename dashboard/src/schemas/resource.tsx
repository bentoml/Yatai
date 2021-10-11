import { IBaseSchema } from './base'

export type ResourceType =
    | 'user'
    | 'user_group'
    | 'organization'
    | 'cluster'
    | 'bento'
    | 'bento_version'
    | 'deployment'
    | 'deployment_snapshot'
    | 'yatai_component'

export interface IResourceSchema extends IBaseSchema {
    name: string
    resource_type: ResourceType
}
