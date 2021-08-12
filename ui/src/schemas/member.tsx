import { IUserSchema } from './user'

export type MemberRole = 'guest' | 'developer' | 'admin'

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
