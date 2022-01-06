import { IBentoWithRepositorySchema } from './bento'
import { IBentoRepositorySchema } from './bento_repository'
import { IModelWithRepositorySchema } from './model'
import { IModelRepositorySchema } from './model_repository'
import { IUserSchema } from './user'

export interface IEventSchema {
    name: string
    operation_name: string
    resource?: IBentoWithRepositorySchema | IBentoRepositorySchema | IModelWithRepositorySchema | IModelRepositorySchema
    updated_at: string
    creator?: IUserSchema
}
