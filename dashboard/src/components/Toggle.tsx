import React from 'react'
import { Checkbox, STYLE_TYPE, CheckboxOverrides } from 'baseui/checkbox'

export interface IToggleProps {
    value?: boolean
    onChange?: (newView: boolean) => void
    overrides?: CheckboxOverrides
}

export default function Toggle({ value, onChange, overrides }: IToggleProps) {
    return (
        <Checkbox
            checked={value}
            overrides={overrides}
            checkmarkType={STYLE_TYPE.toggle_round}
            onChange={(e) => {
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                onChange?.((e.target as any).checked)
            }}
        />
    )
}
