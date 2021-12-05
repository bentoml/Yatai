import { listModelRepositories } from '@/services/model_repository'
import React, { useEffect, useState } from 'react'
import { useQuery } from 'react-query'
import _ from 'lodash'
import { Select } from 'baseui/select'

export interface IModelRepositorySelectorProps {
    value?: string
    onChange?: (newValue: string) => void
}

export default function ModelRepositorySelector({ value, onChange }: IModelRepositorySelectorProps) {
    const [keyword, setKeyword] = useState<string>()
    const [options, setOptions] = useState<{ id: string; label: React.ReactNode }[]>([])
    const modelRepositoriesInfo = useQuery(`listModelRepositories:${keyword}`, () =>
        listModelRepositories({ start: 0, count: 100, search: keyword })
    )
    const handleModelInputChange = _.debounce((term: string) => {
        if (!term) {
            setOptions([])
            return
        }
        setKeyword(term)
    })

    useEffect(() => {
        if (modelRepositoriesInfo.isSuccess) {
            setOptions(
                modelRepositoriesInfo.data?.items.map((m) => ({
                    id: m.name,
                    label: m.name,
                })) ?? []
            )
        } else {
            setOptions([])
        }
    }, [modelRepositoriesInfo.data?.items, modelRepositoriesInfo.isSuccess])

    return (
        <Select
            isLoading={modelRepositoriesInfo.isFetching}
            options={options}
            onChange={(params) => {
                if (!params.option) {
                    return
                }
                onChange?.(params.option.id as string)
            }}
            onInputChange={(e) => {
                const target = e.target as HTMLInputElement
                handleModelInputChange(target.value)
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
