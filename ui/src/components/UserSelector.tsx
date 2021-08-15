import { listUsers } from '@/services/user'
import { Select } from 'baseui/select'
import _ from 'lodash'
import React, { useEffect, useState } from 'react'
import { useQuery } from 'react-query'
import User from './User'

export interface IUserSelectorProps {
    value?: string[]
    onChange?: (newValue: string[]) => void
}

export default function UserSelector({ value, onChange }: IUserSelectorProps) {
    const [keyword, setKeyword] = useState<string | undefined>(undefined)
    const [options, setOptions] = useState<{ id: string; label: React.ReactNode }[]>([])
    const usersInfo = useQuery('listUsers', () => listUsers({ start: 0, count: 100, search: keyword }))

    const handleInputChange = _.debounce((term: string) => {
        if (!term) {
            setOptions([])
            return
        }
        setKeyword(term)
        usersInfo.refetch()
    })
    useEffect(() => {
        if (usersInfo.isSuccess) {
            setOptions(
                usersInfo.data?.items.map((item) => ({
                    id: item.name,
                    label: <User key={item.uid} user={item} />,
                })) ?? []
            )
        } else {
            setOptions([])
        }
    }, [usersInfo.data?.items, usersInfo.isSuccess])

    return (
        <Select
            multi
            isLoading={usersInfo.isFetching}
            options={options}
            onChange={(params) => {
                onChange?.(params.value.map((item) => (item.id as string) ?? '').filter((name) => name !== ''))
            }}
            onInputChange={(e) => {
                const target = e.target as HTMLInputElement
                handleInputChange(target.value)
            }}
            value={value?.map((item) => ({
                id: item,
            }))}
        />
    )
}
