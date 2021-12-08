import { listModels } from '@/services/model'
import { Select } from 'baseui/select'
import _ from 'lodash'
import React, { useState } from 'react'
import { useQuery } from 'react-query'

export interface IModelSelectorProps {
    modelRepositoryName: string
    value?: string
    onChange?: (newValue: string) => void
}

export default function ModelSelector({ modelRepositoryName, value, onChange }: IModelSelectorProps) {
    const [keyword, setKeyword] = useState<string>()
    const [options, setOptions] = useState<{ id: string; label: React.ReactNode }[]>([])
    const modelsInfo = useQuery(`listModels:${modelRepositoryName}:${keyword}`, () => {
        listModels(modelRepositoryName, { start: 0, count: 100, search: keyword })
    })

    const handleModelInputChange = _.debounce((term: string) => {
        if (!term) {
            setOptions([])
            return
        }
        setKeyword(term)
    })

    return (
        <Select
            isLoading={modelsInfo.isLoading}
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
