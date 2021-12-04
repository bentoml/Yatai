import { IBaseSchema } from './base'
import { LabelItemsSchema } from './label'

export type ResourceType =
    | 'user'
    | 'user_group'
    | 'organization'
    | 'cluster'
    | 'bento'
    | 'bento_version'
    | 'deployment'
    | 'deployment_revision'
    | 'yatai_component'
    | 'model'
    | 'model_version'
    | 'api_token'

export interface IResourceSchema extends IBaseSchema {
    name: string
    resource_type: ResourceType
    labels: LabelItemsSchema
}
