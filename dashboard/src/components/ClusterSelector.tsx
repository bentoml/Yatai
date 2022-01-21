import { listClusters } from '@/services/cluster'
import { Select, SelectProps } from 'baseui/select'
import _ from 'lodash'
import React, { useEffect, useState } from 'react'
import { useQuery } from 'react-query'

export interface IClusterSelectorProps {
    value?: string
    onChange?: (newValue: string) => void
    overrides?: SelectProps['overrides']
    disabled?: boolean
}

export default function ClusterSelector({ value, onChange, overrides, disabled }: IClusterSelectorProps) {
    const [keyword, setKeyword] = useState<string>()
    const [options, setOptions] = useState<{ id: string; label: React.ReactNode }[]>([])
    const clustersInfo = useQuery(`listClusters:${keyword}`, () =>
        listClusters({ start: 0, count: 100, search: keyword })
    )

    const handleClusterInputChange = _.debounce((term: string) => {
        if (!term) {
            setOptions([])
            return
        }
        setKeyword(term)
    })

    useEffect(() => {
        if (clustersInfo.isSuccess) {
            setOptions(
                clustersInfo.data?.items.map((item) => ({
                    id: item.name,
                    label: item.name,
                })) ?? []
            )
        } else {
            setOptions([])
        }
    }, [clustersInfo.data?.items, clustersInfo.isSuccess])

    return (
        <Select
            disabled={disabled}
            overrides={overrides}
            isLoading={clustersInfo.isFetching}
            options={options}
            onChange={(params) => {
                if (!params.option) {
                    return
                }
                onChange?.(params.option.id as string)
            }}
            onInputChange={(e) => {
                const target = e.target as HTMLInputElement
                handleClusterInputChange(target.value)
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
