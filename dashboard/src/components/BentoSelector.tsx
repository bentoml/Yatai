import { listBentos } from '@/services/bento'
import { useStyletron } from 'baseui'
import { Select } from 'baseui/select'
import { MonoParagraphXSmall } from 'baseui/typography'
import _ from 'lodash'
import React, { useEffect, useState } from 'react'
import { useQuery } from 'react-query'
import ImageBuildStatusIcon from './ImageBuildStatusIcon'
import Time from './Time'

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
    const [, theme] = useStyletron()

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
                    disabled: item.image_build_status !== 'success',
                    label: (
                        <div
                            style={{
                                display: 'flex',
                                flexDirection: 'row',
                                alignItems: 'center',
                                justifyContent: 'space-between',
                                gap: 42,
                            }}
                        >
                            <div
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: 6,
                                }}
                            >
                                <ImageBuildStatusIcon key={item.uid} size={12} status={item.image_build_status} />
                                <MonoParagraphXSmall
                                    overrides={{
                                        Block: {
                                            style: {
                                                margin: 0,
                                            },
                                        },
                                    }}
                                >
                                    {item.version}
                                </MonoParagraphXSmall>
                            </div>
                            <Time
                                time={item.created_at}
                                style={{
                                    color: theme.colors.contentSecondary,
                                    fontSize: '11px',
                                }}
                            />
                        </div>
                    ),
                })) ?? []
            )
        } else {
            setOptions([])
        }
    }, [bentosInfo.data?.items, bentosInfo.isSuccess, theme.colors.contentSecondary])

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
