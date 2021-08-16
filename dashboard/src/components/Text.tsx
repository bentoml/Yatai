import React from 'react'
import { useStyletron } from 'baseui'

export interface ITextProps {
    children: React.ReactNode
    size?: 'small' | 'medium' | 'large'
    style?: React.CSSProperties
}

const fontSizeMap: { [k in Exclude<ITextProps['size'], undefined>]: string } = {
    small: '12px',
    medium: '13px',
    large: '14px',
}

export default function Text({ children, style, size = 'medium' }: ITextProps) {
    const [css, theme] = useStyletron()
    return (
        <span
            style={style}
            className={css({
                fontSize: fontSizeMap[size],
                color: theme.colors.contentPrimary,
            })}
        >
            {children}
        </span>
    )
}
