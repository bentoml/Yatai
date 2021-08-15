import { Item, Navigation } from 'baseui/side-navigation'
import _ from 'lodash'
import React, { useMemo } from 'react'
import { useLocation, useHistory } from 'react-router-dom'
import useSidebarWidth from '@/hooks/useSidebarWidth'
import { useStyletron } from 'baseui'
import Text from './Text'

export interface IComposedSidebarProps {
    style?: React.CSSProperties
    navStyle?: React.CSSProperties
}

export interface INavItem {
    title: string
    icon?: React.ReactNode
    path: string
    children?: INavItem[]
}

function transformNavItems(navItems: INavItem[]): Item[] {
    return navItems.map((item) => {
        return {
            title: (
                <div
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 8,
                    }}
                >
                    {item.icon} {item.title}
                </div>
            ),
            itemId: item.path,
        }
    })
}

export interface IBaseSideBarProps extends IComposedSidebarProps {
    title?: string
    icon?: React.ReactNode
    navItems: INavItem[]
}

export default function BaseSidebar({ navItems, style, title, icon }: IBaseSideBarProps) {
    const width = useSidebarWidth()

    const history = useHistory()
    const location = useLocation()

    const baseuiNavItems = useMemo(() => transformNavItems(navItems), [navItems])

    const activeItemId = useMemo(() => {
        const item = baseuiNavItems
            .slice()
            .reverse()
            .find((item_) => _.startsWith(location.pathname, item_.itemId))
        return item?.itemId
    }, [baseuiNavItems, location.pathname])

    const [, theme] = useStyletron()

    return (
        <div
            style={{
                width,
                display: 'flex',
                flexDirection: 'column',
                flexBasis: width,
                overflow: 'hidden',
                overflowY: 'auto',
                borderRight: `1px solid ${theme.borders.border200.borderColor}`,
                ...style,
            }}
        >
            {title && icon && (
                <div
                    style={{
                        display: 'flex',
                        gap: 10,
                        fontSize: '11px',
                        // background: theme.colors.backgroundSecondary,
                        alignItems: 'center',
                        padding: '8px 8px 8px 30px',
                        borderBottom: `1px solid ${theme.borders.border200.borderColor}`,
                        overflow: 'hidden',
                    }}
                >
                    {icon}
                    <Text
                        style={{
                            fontSize: '12px',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap',
                        }}
                    >
                        {title}
                    </Text>
                </div>
            )}
            <Navigation
                overrides={{
                    Root: {
                        style: {
                            fontSize: '14px',
                        },
                    },
                }}
                activeItemId={activeItemId ?? baseuiNavItems[0].itemId ?? ''}
                items={baseuiNavItems}
                onChange={({ event, item }) => {
                    event.preventDefault()
                    history.push(item.itemId)
                }}
            />
        </div>
    )
}
