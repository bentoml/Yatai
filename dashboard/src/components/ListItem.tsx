import React from 'react'
import { ListItem as BaseListItem, PropsT } from 'baseui/list'
import { useStyletron } from 'baseui'
import classNames from 'classnames'

export interface IListItemProps extends PropsT {
    className?: string
    onClick?: (event: React.MouseEvent<HTMLDivElement, MouseEvent>) => void
    style?: React.CSSProperties
}

export default function ListItem({ className, style, onClick, ...props }: IListItemProps) {
    const [, theme] = useStyletron()
    return (
        <div className={classNames('listItem', className)} style={style} onClick={onClick} role='button' tabIndex={0}>
            <BaseListItem
                overrides={{
                    Root: {
                        style: {
                            'cursor': 'pointer',
                            ':hover': {
                                backgroundColor: theme.colors.backgroundSecondary,
                            },
                            // eslint-disable-next-line
                            ...(((props.overrides?.Root as any)?.style as any) ?? {}),
                        },
                    },
                }}
                // eslint-disable-next-line react/jsx-props-no-spreading
                {...props}
            />
        </div>
    )
}
