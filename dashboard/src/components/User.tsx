import React from 'react'
import { IUserSchema } from '@/schemas/user'
import { Avatar } from 'baseui/avatar'
import Text from './Text'

export interface IUserProps {
    user: IUserSchema
    apiTokenName?: string
    size?: 'scale700' | 'scale800' | 'scale1000' | 'scale1200' | 'scale1400' | '64px' | '32px' | '16px' | '14px'
    style?: React.CSSProperties
}

export default function User({ user, apiTokenName, size = 'scale800', style }: IUserProps) {
    const name = !user.first_name && !user.last_name ? user.name : `${user.first_name} ${user.last_name}`

    const displayName = apiTokenName ? `${name} (${apiTokenName})` : name

    return (
        <div
            style={{
                display: 'flex',
                alignItems: 'center',
                gap: 10,
                ...style,
            }}
        >
            <Avatar size={size} name={name} src={user.avatar_url} />
            <Text>{displayName}</Text>
        </div>
    )
}
