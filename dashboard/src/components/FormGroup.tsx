/* eslint-disable react/jsx-props-no-spreading */
import React from 'react'
import type { IconBaseProps } from 'react-icons/lib'

interface IFormGroupProps {
    children: React.ReactNode
    icon?: React.ComponentType<IconBaseProps>
    style?: React.CSSProperties
}

export default function FormGroup({ children, icon, style }: IFormGroupProps) {
    return (
        <div
            style={{
                display: 'flex',
                flexDirection: 'row',
                justifyContent: 'center',
                marginBottom: 10,
                gap: 16,
                ...style,
            }}
        >
            <div style={{ paddingTop: 6 }}>{icon && React.createElement(icon, { size: 20 })}</div>
            <div style={{ flexGrow: 1 }}>{children}</div>
        </div>
    )
}
