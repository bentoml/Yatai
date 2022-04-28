import { IBaseSchema } from './base'
import { LabelItemsSchema } from './label'

export type ResourceType =
    | 'user'
    | 'user_group'
    | 'organization'
    | 'cluster'
    | 'bento_repository'
    | 'bento'
    | 'deployment'
    | 'deployment_revision'
    | 'yatai_component'
    | 'model_repository'
    | 'model'
    | 'api_token'
    | 'bento_runner'
    | 'bento_api_server'

export interface IResourceSchema extends IBaseSchema {
    name: string
    resource_type: ResourceType
    labels: LabelItemsSchema
}
