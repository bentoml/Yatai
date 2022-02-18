import { IUserSchema } from './user'
import { MemberRole } from './member_role'

export interface IMemberSchema {
    user: IUserSchema
    role: MemberRole
}

export interface ICreateMembersSchema {
    usernames: string[]
    role: MemberRole
}

export interface IDeleteMemberSchema {
    username: string
}
