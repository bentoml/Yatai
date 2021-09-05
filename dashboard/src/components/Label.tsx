import { useStyletron } from 'baseui'
import React from 'react'

export interface ILabelProps {
    children: React.ReactNode
    style?: React.CSSProperties
}

export default function Label({ children, style }: ILabelProps) {
    const [, theme] = useStyletron()

    return (
        // eslint-disable-next-line jsx-a11y/label-has-associated-control
        <label
            style={{
                fontWeight: 500,
                color: theme.colors.contentPrimary,
                ...style,
            }}
        >
            {children}
        </label>
    )
}
