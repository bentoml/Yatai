/* eslint-disable jsx-a11y/no-static-element-interactions */
import React, { useCallback } from 'react'
import classNames from 'classnames'
import { useStyletron } from 'baseui'
import { Skeleton } from 'baseui/skeleton'
import { createUseStyles } from 'react-jss'
import { Theme } from 'baseui/theme'
import { IThemedStyleProps } from '@/interfaces/IThemedStyle'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import Text from '@/components/Text'
import type { IconBaseProps } from 'react-icons/lib'

import styles from './index.module.scss'

const getLinkStyle = (theme: Theme) => {
    return {
        color: theme.colors.contentPrimary,
    }
}

const useStyles = createUseStyles({
    card: (props: IThemedStyleProps) => {
        const linkStyle = getLinkStyle(props.theme)

        return {
            'box-shadow': props.theme.lighting.shadow400,
            'background': props.theme.colors.backgroundPrimary,
            '& a': linkStyle,
            '& a:link': linkStyle,
            '& a:visited': linkStyle,
        }
    },
})

export interface ICardProps {
    title?: string | React.ReactNode
    onTitleChange?: (title: string) => Promise<void>
    titleIcon?: React.ComponentType<IconBaseProps>
    titleTail?: React.ReactNode
    style?: React.CSSProperties
    headStyle?: React.CSSProperties
    bodyStyle?: React.CSSProperties
    bodyClassName?: string
    children?: React.ReactNode
    className?: string
    extra?: React.ReactNode
    loading?: boolean
    onMountCard?: React.RefCallback<HTMLDivElement>
    onClick?: () => void
}

export default function Card({
    title,
    titleIcon: TitleIcon,
    titleTail,
    extra,
    className,
    style,
    headStyle,
    bodyStyle,
    bodyClassName,
    children,
    loading,
    onMountCard,
    onClick,
}: ICardProps) {
    const mountCard = useCallback(
        (card) => {
            if (card) {
                // eslint-disable-next-line no-param-reassign
                card.style.transform = 'translate3d(0, 0, 0)'
                onMountCard?.(card)
            }
        },
        // eslint-disable-next-line react-hooks/exhaustive-deps
        []
    )

    let c = children
    if (loading) {
        c = <Skeleton rows={3} animation />
    }

    const [, theme] = useStyletron()

    const themeType = useCurrentThemeType()

    const dynamicStyles = useStyles({ theme, themeType })

    return (
        <div
            ref={mountCard}
            onClick={onClick}
            className={classNames(styles.card, dynamicStyles.card, className)}
            style={style}
        >
            {(title || extra) && (
                <div
                    className={styles.cardHead}
                    style={{
                        ...headStyle,
                        color: theme.colors.contentPrimary,
                        borderBottomColor: theme.borders.border300.borderColor,
                    }}
                >
                    <div className={styles.cardHeadWrapper}>
                        {title && (
                            <div className={styles.cardHeadTitle}>
                                {TitleIcon && <TitleIcon size={12} />}
                                {typeof title === 'string' ? <Text size='large'>{title}</Text> : title}
                                {titleTail}
                            </div>
                        )}
                        {extra && <div className={styles.cardExtra}>{extra}</div>}
                    </div>
                </div>
            )}
            <div className={classNames(styles.cardBody, bodyClassName)} style={bodyStyle}>
                {c}
            </div>
        </div>
    )
}
