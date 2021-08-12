import { IBaseSchema } from './base'

export type ResourceType = 'user' | 'organization' | 'cluster' | 'bundle' | 'bundle_version' | 'deployment'

export interface IResourceSchema extends IBaseSchema {
    name: string
    resource_type: ResourceType
}
