import React, { useCallback, useEffect, useState } from 'react'
import { ILokiLineFilterNode, LokiFilterType } from '@/interfaces/ILoki'
import useTranslation from '@/hooks/useTranslation'
import { Button } from 'baseui/button'
import { DeleteAlt } from 'baseui/icon'
import { Select } from 'baseui/select'
import { Input } from 'baseui/input'
import Toggle from './Toggle'
import Label from './Label'

interface ILokiFiltersFormProps {
    filters: ILokiLineFilterNode[]
    onSubmit: (filters: ILokiLineFilterNode[]) => void
}

export default ({ filters, onSubmit }: ILokiFiltersFormProps) => {
    const [currentFilters, setCurrentFilters] = useState(filters)

    useEffect(() => {
        setCurrentFilters(filters)
    }, [filters])

    const handleUpdate = useCallback((filter: ILokiLineFilterNode, idx: number) => {
        setCurrentFilters((filters_) =>
            filters_.map((f, i) => {
                if (i === idx) {
                    return filter
                }
                return f
            })
        )
    }, [])

    const handleDelete = useCallback((idx: number) => {
        setCurrentFilters((filters_) =>
            filters_.filter((_f, i) => {
                return i !== idx
            })
        )
    }, [])

    const [t] = useTranslation()

    return (
        <div>
            <ul
                style={{
                    padding: 0,
                    margin: 0,
                }}
            >
                {currentFilters.map((filter, idx) => {
                    return (
                        <li
                            key={idx}
                            style={{
                                display: 'flex',
                                flexDirection: 'row',
                                alignItems: 'center',
                                marginBottom: 20,
                                gap: 10,
                            }}
                        >
                            <Button
                                overrides={{
                                    Root: {
                                        style: {
                                            flexShrink: 0,
                                            marginRight: 20,
                                        },
                                    },
                                }}
                                size='mini'
                                shape='circle'
                                onClick={(e) => {
                                    e.preventDefault()
                                    handleDelete(idx)
                                }}
                            >
                                <DeleteAlt />
                            </Button>
                            <Label
                                style={{
                                    flexShrink: 0,
                                }}
                            >
                                {t('regexp')}
                            </Label>
                            <Toggle
                                overrides={{
                                    Root: {
                                        style: {
                                            marginBottom: 0,
                                            marginRight: 10,
                                        },
                                    },
                                }}
                                value={filter.isRegexp}
                                onChange={(v) => {
                                    handleUpdate(
                                        {
                                            ...filter,
                                            isRegexp: v,
                                        },
                                        idx
                                    )
                                }}
                            />
                            <Select
                                size='compact'
                                clearable={false}
                                searchable={false}
                                overrides={{
                                    Root: {
                                        style: {
                                            marginRight: 10,
                                            width: '200px',
                                            flexShrink: 0,
                                        },
                                    },
                                }}
                                value={[{ id: filter.type }]}
                                onChange={({ option }) => {
                                    handleUpdate(
                                        {
                                            ...filter,
                                            type: option?.id as LokiFilterType,
                                        },
                                        idx
                                    )
                                }}
                                options={[
                                    {
                                        id: 'contains' as LokiFilterType,
                                        label: t('contains'),
                                    },
                                    {
                                        id: 'not contains' as LokiFilterType,
                                        label: t('not contains'),
                                    },
                                ]}
                            />
                            <Input
                                size='compact'
                                overrides={{
                                    Root: {
                                        style: {
                                            flexGrow: 1,
                                        },
                                    },
                                }}
                                value={filter.value}
                                onChange={(e) => {
                                    const v = (e.target as HTMLInputElement).value
                                    handleUpdate(
                                        {
                                            ...filter,
                                            value: v || '',
                                        },
                                        idx
                                    )
                                }}
                            />
                        </li>
                    )
                })}
            </ul>
            <div
                style={{
                    display: 'flex',
                    marginBottom: 20,
                    gap: 8,
                }}
            >
                <Button
                    size='compact'
                    overrides={{
                        Root: {
                            style: {
                                borderStyle: 'dashed',
                                flexGrow: 1,
                            },
                        },
                    }}
                    kind='tertiary'
                    onClick={() => {
                        setCurrentFilters([])
                    }}
                >
                    {t('clear')}
                </Button>
                <Button
                    size='compact'
                    kind='secondary'
                    overrides={{
                        Root: {
                            style: {
                                borderStyle: 'dashed',
                                flexGrow: 6,
                            },
                        },
                    }}
                    onClick={() => {
                        setCurrentFilters((filters_) => [
                            ...filters_,
                            {
                                type: 'contains',
                                isRegexp: false,
                                value: '',
                            },
                        ])
                    }}
                >
                    {t('add filter condition')}
                </Button>
            </div>
            <div
                style={{
                    display: 'flex',
                }}
            >
                <div
                    style={{
                        flexGrow: 1,
                    }}
                />
                <Button
                    size='compact'
                    onClick={() => {
                        onSubmit(currentFilters)
                    }}
                >
                    {t('submit')}
                </Button>
            </div>
        </div>
    )
}
