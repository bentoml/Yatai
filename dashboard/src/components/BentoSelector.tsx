import { listBentos } from '@/services/bento'
import { Select } from 'baseui/select'
import _ from 'lodash'
import React, { useEffect, useState } from 'react'
import { useQuery } from 'react-query'
import BentoImageBuildStatusTag from './BentoImageBuildStatus'

export interface IBentoSelectorProps {
    bentoRepositoryName: string
    value?: string
    onChange?: (newValue: string) => void
}

export default function BentoSelector({ bentoRepositoryName, value, onChange }: IBentoSelectorProps) {
    const [keyword, setKeyword] = useState<string>()
    const [options, setOptions] = useState<{ id: string; label: React.ReactNode }[]>([])
    const bentosInfo = useQuery(`listBento:${bentoRepositoryName}:${keyword}`, () =>
        listBentos(bentoRepositoryName, { start: 0, count: 100, search: keyword })
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
                    id: item.version,
                    label: (
                        <div
                            style={{
                                display: 'flex',
                                flexDirection: 'row',
                                alignItems: 'center',
                                gap: 10,
                            }}
                        >
                            {item.version}
                            <BentoImageBuildStatusTag key={item.uid} status={item.image_build_status} />
                        </div>
                    ),
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
