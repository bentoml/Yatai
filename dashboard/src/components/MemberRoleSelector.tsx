import React from 'react'
import useTranslation from '@/hooks/useTranslation'
import { MemberRole } from '@/schemas/member'
import { Select } from 'baseui/select'

export interface IMemberRoleSelectorProps {
    value?: MemberRole
    onChange?: (newValue: MemberRole) => void
}

export default function MemberRoleSelector({ value, onChange }: IMemberRoleSelectorProps) {
    const [t] = useTranslation()

    return (
        <Select
            options={
                [
                    {
                        id: 'guest',
                        label: t('guest'),
                    },
                    {
                        id: 'developer',
                        label: t('developer'),
                    },
                    {
                        id: 'admin',
                        label: t('admin'),
                    },
                ] as { id: MemberRole; label: string }[]
            }
            onChange={(params) => {
                if (!params.option) {
                    return
                }
                onChange?.(params.option.id as MemberRole)
            }}
            value={
                value
                    ? [
                          {
                              id: value,
                          },
                      ]
                    : []
            }
        />
    )
}
