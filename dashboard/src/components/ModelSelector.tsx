import { listModels } from '@/services/model'
import React, { useEffect, useState } from 'react'
import { useQuery } from 'react-query'
import _ from 'lodash'
import { Select } from 'baseui/select'

export interface IModelSelectorProps {
    value?: string
    onChange?: (newValue: string) => void
}

export default function ModelSelector({ value, onChange }: IModelSelectorProps) {
    const [keyword, setKeyword] = useState<string>()
    const [options, setOptions] = useState<{ id: string; label: React.ReactNode }[]>([])
    const modelsInfo = useQuery(`listModels:${keyword}`, () => listModels({ start: 0, count: 100, search: keyword }))
    const handleModelInputChange = _.debounce((term: string) => {
        if (!term) {
            setOptions([])
            return
        }
        setKeyword(term)
    })

    useEffect(() => {
        if (modelsInfo.isSuccess) {
            setOptions(
                modelsInfo.data?.items.map((m) => ({
                    id: m.name,
                    label: m.name,
                })) ?? []
            )
        } else {
            setOptions([])
        }
    }, [modelsInfo.data?.items, modelsInfo.isSuccess])

    return (
        <Select
            isLoading={modelsInfo.isFetching}
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
