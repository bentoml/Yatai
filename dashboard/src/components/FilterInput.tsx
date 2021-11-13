import { useQ } from '@/hooks/useQ'
import useTranslation from '@/hooks/useTranslation'
import { parseQ, qToString } from '@/utils'
import { Button } from 'baseui/button'
import { Input } from 'baseui/input'
import React, { useEffect, useMemo, useState } from 'react'
import { AiOutlineSearch } from 'react-icons/ai'
import Filter from './Filter'

export interface IFilterCondition {
    label: React.ReactNode
    qStr: string
}

export interface IFilterInputProps {
    filterConditions?: IFilterCondition[]
}

export default function FilterInput({ filterConditions }: IFilterInputProps) {
    const { q, replaceQ } = useQ()
    const qStr = useMemo(() => {
        const s = qToString(q)
        if (s) {
            return `${s} `
        }
        return s
    }, [q])
    const [value, setValue] = useState(qStr)
    useEffect(() => {
        setValue(qStr)
    }, [qStr])
    const [t] = useTranslation()
    const isKeyDownRef = React.useRef(false)
    useEffect(() => {
        if (isKeyDownRef.current) {
            return
        }
        replaceQ(parseQ(value))
    }, [replaceQ, value])

    return (
        <div
            style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
            }}
        >
            {filterConditions && (
                <div
                    style={{
                        fontWeight: 'normal',
                    }}
                >
                    <Filter
                        label={t('filters')}
                        options={filterConditions.map((item) => {
                            return {
                                id: item.qStr,
                                label: item.label,
                            }
                        })}
                        onChange={({ value: value_ }) => {
                            const s = value_.map((v) => v.id).join(' ')
                            replaceQ(parseQ(s))
                        }}
                        value={[
                            {
                                id: qToString(q),
                            },
                        ]}
                    />
                </div>
            )}
            <div
                style={{
                    flexGrow: 1,
                    flexShrink: 0,
                }}
            >
                <Input
                    startEnhancer={<AiOutlineSearch size={12} />}
                    clearable
                    size='compact'
                    value={value}
                    onChange={(e) => {
                        setValue(e.currentTarget.value)
                    }}
                    onKeyUp={() => {
                        isKeyDownRef.current = false
                    }}
                    onKeyDown={(e) => {
                        isKeyDownRef.current = true
                        if (e.key === 'Enter') {
                            const q_ = parseQ(value)
                            replaceQ(q_)
                        }
                    }}
                />
            </div>
            <div>
                <Button
                    startEnhancer={<AiOutlineSearch size={12} />}
                    kind='tertiary'
                    size='compact'
                    onClick={() => replaceQ(parseQ(value))}
                >
                    {t('search')}
                </Button>
            </div>
        </div>
    )
}
