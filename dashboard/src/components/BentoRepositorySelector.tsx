import { listBentoRepositories } from '@/services/bento_repository'
import { Select } from 'baseui/select'
import _ from 'lodash'
import React, { useEffect, useState } from 'react'
import { useQuery } from 'react-query'

export interface IBentoRepositorySelectorProps {
    value?: string
    onChange?: (newValue: string) => void
}

export default function BentoRepositorySelector({ value, onChange }: IBentoRepositorySelectorProps) {
    const [keyword, setKeyword] = useState<string>()
    const [options, setOptions] = useState<{ id: string; label: React.ReactNode }[]>([])
    const bentosInfo = useQuery(`listBentoRepositories:${keyword}`, () =>
        listBentoRepositories({ start: 0, count: 100, search: keyword })
    )

    const handleBentoInputChange = _.debounce((term: string) => {
        if (!term) {
            setOptions([])
            return
        }
        setKeyword(term)
    })

    useEffect(() => {
        if (bentosInfo.isSuccess) {
            setOptions(
                bentosInfo.data?.items.map((item) => ({
                    id: item.name,
                    label: item.name,
                })) ?? []
            )
        } else {
            setOptions([])
        }
    }, [bentosInfo.data?.items, bentosInfo.isSuccess])

    return (
        <Select
            isLoading={bentosInfo.isFetching}
            options={options}
            onChange={(params) => {
                if (!params.option) {
                    return
                }
                onChange?.(params.option.id as string)
            }}
            onInputChange={(e) => {
                const target = e.target as HTMLInputElement
                handleBentoInputChange(target.value)
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
