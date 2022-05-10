import React from 'react'
import { Checkbox, STYLE_TYPE, CheckboxOverrides, LABEL_PLACEMENT } from 'baseui/checkbox'

export interface IToggleProps {
    value?: boolean
    onChange?: (newView: boolean) => void
    overrides?: CheckboxOverrides
    disabled?: boolean
    children?: React.ReactNode
    labelPlacement?: keyof LABEL_PLACEMENT
}

export default function Toggle({ value, onChange, overrides, disabled, children, labelPlacement }: IToggleProps) {
    return (
        <Checkbox
            disabled={disabled}
            checked={value}
            overrides={overrides}
            checkmarkType={STYLE_TYPE.toggle_round}
            onChange={(e) => {
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                onChange?.((e.target as any).checked)
            }}
            labelPlacement={labelPlacement}
        >
            {children}
        </Checkbox>
    )
}
