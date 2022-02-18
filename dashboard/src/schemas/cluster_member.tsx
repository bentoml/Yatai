import { IBaseSchema } from './base'
import { IClusterSchema } from './cluster'
import { MemberRole } from './member_role'
import { IUserSchema } from './user'

export interface IClusterMemberSchema extends IBaseSchema {
    user: IUserSchema
    cluster: IClusterSchema
    role: MemberRole
    creator?: IUserSchema
}
