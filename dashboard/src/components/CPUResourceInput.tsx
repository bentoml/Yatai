import React from 'react'
import { strToCPUMilliCores } from '@/utils'
import { Input } from 'baseui/input'

interface ICPUResourceInputProps {
    value?: string
    onChange?: (value: string) => void
}

export const CPUResourceInput = ({ value, onChange }: ICPUResourceInputProps) => {
    const milliCores = strToCPUMilliCores(value)
    const n = milliCores === 0 ? '' : String(milliCores)

    return (
        <Input
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
