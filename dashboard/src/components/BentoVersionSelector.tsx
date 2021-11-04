import { listBentoVersions } from '@/services/bento_version'
import { Select } from 'baseui/select'
import _ from 'lodash'
import React, { useEffect, useState } from 'react'
import { useQuery } from 'react-query'
import BentoVersionImageBuildStatusTag from './BentoVersionImageBuildStatus'

export interface IBentoVersionSelectorProps {
    bentoName: string
    value?: string
    onChange?: (newValue: string) => void
}

export default function BentoVersionSelector({ bentoName, value, onChange }: IBentoVersionSelectorProps) {
    const [keyword, setKeyword] = useState<string>()
    const [options, setOptions] = useState<{ id: string; label: React.ReactNode }[]>([])
    const bentosInfo = useQuery(`listBentoVersions:${keyword}`, () =>
        listBentoVersions(bentoName, { start: 0, count: 100, search: keyword })
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
                            <BentoVersionImageBuildStatusTag key={item.uid} status={item.image_build_status} />
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
