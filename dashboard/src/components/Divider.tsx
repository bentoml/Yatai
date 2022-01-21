import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import { IThemedStyleProps } from '@/interfaces/IThemedStyle'
import { useStyletron } from 'baseui'
import classNames from 'classnames'
import React from 'react'
import { createUseStyles } from 'react-jss'

const useStyles = createUseStyles({
    wrapper: (props: IThemedStyleProps) => {
        return {
            'margin': '28px 0',
            'fontSize': '18px',
            'display': 'flex',
            'alignItems': 'center',
            'justifyContent': 'center',
            '&:before': {
                content: '""',
                position: 'relative',
                top: '50%',
                borderTop: `1px solid ${props.theme.borders.border300.borderColor}`,
                transform: 'translateY(50%)',
            },
            '&:after': {
                content: '""',
                position: 'relative',
                top: '50%',
                borderTop: `1px solid ${props.theme.borders.border300.borderColor}`,
                transform: 'translateY(50%)',
            },
        }
    },
    center: {
        '&:before': {
            width: '50%',
        },
        '&:after': {
            width: '50%',
        },
    },
    left: {
        '&:before': {
            width: '0%',
        },
        '&:after': {
            width: '100%',
        },
        '& $innerText': {
            paddingLeft: 0,
        },
    },
    right: {
        '&:before': {
            width: '100%',
        },
        '&:after': {
            width: '0%',
        },
        '& $innerText': {
            paddingRight: 0,
        },
    },
    innerText: {
        flexShrink: 0,
        fontWeight: '500',
        padding: '0 1em',
    },
})

export interface IDividerProps {
    children: React.ReactNode
    orientation?: 'left' | 'center' | 'right'
}

export default function Divider({ children, orientation = 'center' }: IDividerProps) {
    const themeType = useCurrentThemeType()
    const [, theme] = useStyletron()
    const styles = useStyles({ themeType, theme })
    return (
        <div className={classNames(styles.wrapper, styles[orientation])}>
            <div className={styles.innerText}>{children}</div>
        </div>
    )
}
