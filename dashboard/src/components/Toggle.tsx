import React from 'react'
import { Checkbox, STYLE_TYPE } from 'baseui/checkbox'

export interface IToggleProps {
    value?: boolean
    onChange?: (newView: boolean) => void
}

export default function Toggle({ value, onChange }: IToggleProps) {
    return (
        <Checkbox
            checked={value}
            checkmarkType={STYLE_TYPE.toggle_round}
            onChange={(e) => {
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                onChange?.((e.target as any).checked)
            }}
        />
    )
}
