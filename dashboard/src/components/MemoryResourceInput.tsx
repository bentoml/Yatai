import { Input, InputProps } from 'baseui/input'
import { Select } from 'baseui/select'
import React, { useRef } from 'react'

interface IMemoryResourceInputProps {
    value?: string
    onChange?: (value: string) => void
    overrides?: InputProps['overrides']
}

export default function MemoryResourceInput({ value, onChange, overrides }: IMemoryResourceInputProps) {
    const unitRef = useRef('Mi')
    const vRef = useRef<number | undefined>(undefined)
    if (value) {
        const m = value.match(/(\d+)(Ki|Mi|Gi|Ti|Pi|Ei)/)
        if (m) {
            vRef.current = parseInt(m[1], 10)
            // eslint-disable-next-line prefer-destructuring
            unitRef.current = m[2]
        }
    }

    return (
        <div
            style={{
                display: 'flex',
                alignItems: 'center',
            }}
        >
            <Input
                overrides={overrides}
                type='number'
                min={0}
                value={String(vRef.current)}
                onChange={(e) => {
                    const v = (e.target as HTMLInputElement).value
                    if (!unitRef.current || v === undefined) {
                        return
                    }
                    vRef.current = parseInt(v, 10)
                    if (v === '0' || !v) {
                        onChange?.('')
                        return
                    }
                    onChange?.(`${v}${unitRef.current}`)
                }}
            />
            <Select
                clearable={false}
                searchable={false}
                overrides={{
                    Root: {
                        style: {
                            width: '90px',
                        },
                    },
                }}
                options={[
                    {
                        id: 'Ki',
                        label: 'Ki',
                    },
                    {
                        id: 'Mi',
                        label: 'Mi',
                    },
                    {
                        id: 'Gi',
                        label: 'Gi',
                    },
                    {
                        id: 'Ti',
                        label: 'Ti',
                    },
                    {
                        id: 'Pi',
                        label: 'Pi',
                    },
                ]}
                onChange={(params) => {
                    if (!params.option) {
                        return
                    }
                    unitRef.current = String(params.option.id)
                    if (vRef.current === 0) {
                        onChange?.('')
                        return
                    }
                    onChange?.(`${vRef.current}${unitRef.current}`)
                }}
                value={[{ id: unitRef.current }]}
            />
        </div>
    )
}
