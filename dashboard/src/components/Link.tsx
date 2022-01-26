import React from 'react'
import { Link as RouterLink } from 'react-router-dom'
import { StyledLink as BaseLink } from 'baseui/link'
import { createUseStyles } from 'react-jss'
import { IThemedStyleProps } from '@/interfaces/IThemedStyle'
import { useStyletron } from 'baseui'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import _ from 'lodash'

const useStyles = createUseStyles({
    wrapper: (props: IThemedStyleProps) => ({
        'color': props.theme.colors.primary,
        'textDecoration': 'none',
        '& a': {
            textDecoration: 'none',
        },
        '&:hover': {
            textDecoration: 'underline',
        },
        '& a:hover': {
            textDecoration: 'underline',
        },
        '.listItem:hover &': {
            textDecoration: 'underline',
        },
    }),
})

export interface ILinkProps {
    target?: '_blank' | '_self' | '_parent' | '_top'
    href: string
    children: React.ReactNode
    style?: React.CSSProperties
}

export default function Link({ target = '_self', href, children, style }: ILinkProps) {
    const [, theme] = useStyletron()
    const themeType = useCurrentThemeType()

    const styles = useStyles({ theme, themeType })
    const outsideLink = _.startsWith(href, 'http://') || _.startsWith(href, 'https://')

    return outsideLink ? (
        <BaseLink className={styles.wrapper} href={href} target={target}>
            {children}
        </BaseLink>
    ) : (
        <RouterLink
            onClick={(e) => {
                e.stopPropagation()
            }}
            className={styles.wrapper}
            style={style}
            to={href}
            target={target}
        >
            {children}
        </RouterLink>
    )
}
