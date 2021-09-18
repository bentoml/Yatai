import React from 'react'
import { LazyLog as BaseLazyLog, LazyLogProps as BaseLazyLogProps } from 'react-lazylog'
import { createUseStyles } from 'react-jss'
import { IThemedStyleProps } from '@/interfaces/IThemedStyle'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import { useStyletron } from 'baseui'

const useStyles = createUseStyles({
    line: (props: IThemedStyleProps) => ({
        'background': props.theme.colors.backgroundPrimary,
        'color': props.theme.colors.contentPrimary,
        '&:hover': {
            background: props.theme.colors.backgroundPrimary,
        },
        'cursor': 'text',
        'user-select': 'initial',
    }),
})

interface ILazyLogProps extends Partial<BaseLazyLogProps> {
    width?: number | 'auto'
    height?: number | string
}

export default ({ width, height, ...restProps }: ILazyLogProps) => {
    const [, theme] = useStyletron()
    const themeType = useCurrentThemeType()
    const styles = useStyles({ theme, themeType })

    let logContainerStyle: React.CSSProperties = {
        background: theme.colors.backgroundPrimary,
    }

    if (width !== 'auto') {
        logContainerStyle = {
            ...logContainerStyle,
            width,
        }
    }

    // eslint-disable-next-line react/jsx-props-no-spreading
    return <BaseLazyLog height={height} lineClassName={styles.line} style={logContainerStyle} {...restProps} />
}
