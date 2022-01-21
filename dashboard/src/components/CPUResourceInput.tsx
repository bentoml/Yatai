import React from 'react'
import { strToCPUMilliCores } from '@/utils'
import { Input, InputProps } from 'baseui/input'

interface ICPUResourceInputProps {
    value?: string
    onChange?: (value: string) => void
    overrides?: InputProps['overrides']
}

export const CPUResourceInput = ({ value, onChange, overrides }: ICPUResourceInputProps) => {
    const milliCores = strToCPUMilliCores(value)
    const n = milliCores === 0 ? '' : String(milliCores)

    return (
        <Input
            overrides={overrides}
            type='number'
            value={n}
            min={0}
            onChange={(e) => {
                const v = (e.target as HTMLInputElement).value
                if (!v || v === '0') {
                    onChange?.('')
                    return
                }
                onChange?.(`${v}m`)
            }}
            endEnhancer='m'
        />
    )
}
