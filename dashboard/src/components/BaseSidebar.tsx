import { Item, Navigation } from 'baseui/side-navigation'
import _ from 'lodash'
import React, { useCallback, useContext, useMemo } from 'react'
import { useLocation, useHistory } from 'react-router-dom'
import useSidebarWidth from '@/hooks/useSidebarWidth'
import { useStyletron } from 'baseui'
import type { IconBaseProps } from 'react-icons/lib'
import { SidebarContext } from '@/contexts/SidebarContext'
import { sidebarExpandedWidth, sidebarFoldedWidth } from '@/consts'
import { AiOutlineSetting, AiOutlineDoubleLeft, AiOutlineDoubleRight } from 'react-icons/ai'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import color from 'color'
import Text from '@/components/Text'
import useTranslation from '@/hooks/useTranslation'

export interface IComposedSidebarProps {
    style?: React.CSSProperties
    navStyle?: React.CSSProperties
}

export interface INavItem {
    title: string
    icon?: React.ComponentType<IconBaseProps>
    path: string
    children?: INavItem[]
}

function transformNavItems(navItems: INavItem[], expanded = true): Item[] {
    return navItems.map((item) => {
        const { icon: Icon } = item
        return {
            title: (
                <div
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 12,
                        lineHeight: '24px',
                        height: 24,
                    }}
                >
                    {Icon && <Icon size={12} />}
                    {expanded && <span>{item.title}</span>}
                </div>
            ),
            itemId: item.path,
        }
    })
}

export interface IBaseSideBarProps extends IComposedSidebarProps {
    title?: string
    icon?: React.ComponentType<IconBaseProps>
    navItems: INavItem[]
    settingsPath?: string
}

export default function BaseSidebar({ navItems, style, title, icon, settingsPath }: IBaseSideBarProps) {
    const width = useSidebarWidth()
    const ctx = useContext(SidebarContext)

    const history = useHistory()
    const location = useLocation()

    const baseuiNavItems = useMemo(() => transformNavItems(navItems, ctx.expanded), [ctx.expanded, navItems])

    const checkIsSettingsPage = useCallback(
        (currentPath: string) => {
            if (!settingsPath) {
                return false
            }
            return _.startsWith(currentPath, settingsPath)
        },
        [settingsPath]
    )

    const activeItemId = useMemo(() => {
        if (checkIsSettingsPage(location.pathname)) {
            return undefined
        }
        const item = baseuiNavItems
            .slice()
            .reverse()
            .find((item_) => _.startsWith(location.pathname, item_.itemId))
        return item?.itemId
    }, [baseuiNavItems, checkIsSettingsPage, location.pathname])

    const [, theme] = useStyletron()

    const handleExpandedClick = useCallback(() => {
        if (ctx.expanded) {
            ctx.setExpanded(false)
        } else {
            ctx.setExpanded(true)
        }
    }, [ctx])

    const isSettingsPage = checkIsSettingsPage(location.pathname)
    const themeType = useCurrentThemeType()

    const settingNavActiveBackground =
        themeType === 'light'
            ? color(theme.colors.background).darken(0.09).rgb().string()
            : color(theme.colors.background).lighten(0.3).rgb().string()

    const [t] = useTranslation()

    return (
        <div
            style={{
                width,
                display: 'flex',
                flexShrink: 0,
                flexDirection: 'column',
                flexBasis: width,
                overflow: 'hidden',
                overflowY: 'auto',
                background: theme.colors.backgroundPrimary,
                borderRight: `1px solid ${theme.borders.border200.borderColor}`,
                transition: 'all 200ms cubic-bezier(0.7, 0.1, 0.33, 1) 0ms',
                ...style,
            }}
        >
            {ctx.expanded && title && icon && (
                <div
                    style={{
                        display: 'flex',
                        gap: 14,
                        fontSize: '11px',
                        alignItems: 'center',
                        padding: '8px 8px 8px 30px',
                        borderBottom: `1px solid ${theme.borders.border200.borderColor}`,
                        overflow: 'hidden',
                    }}
                >
                    {React.createElement(icon, { size: 10 })}
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
                            flexGrow: 1,
                        },
                    },
                }}
                activeItemId={activeItemId ?? (!isSettingsPage ? baseuiNavItems[0].itemId : '') ?? ''}
                items={baseuiNavItems}
                onChange={({ event, item }) => {
                    event.preventDefault()
                    history.push(item.itemId)
                }}
            />
            <div
                style={{
                    display: 'flex',
                    flexDirection: ctx.expanded ? 'row' : 'column',
                }}
            >
                {settingsPath ? (
                    <div
                        role='button'
                        tabIndex={0}
                        style={{
                            flexGrow: 1,
                            display: 'flex',
                            alignItems: 'center',
                            flexDirection: 'row',
                            height: 48,
                            cursor: 'pointer',
                            transition: 'all 250ms cubic-bezier(0.7, 0.1, 0.33, 1) 0ms',
                            width: ctx.expanded ? sidebarExpandedWidth - sidebarFoldedWidth : sidebarFoldedWidth,
                            overflow: 'hidden',
                            borderLeftWidth: 4,
                            borderLeftStyle: 'solid',
                            borderLeftColor: isSettingsPage ? theme.colors.primary : 'transparent',
                            background: isSettingsPage ? settingNavActiveBackground : 'transparent',
                        }}
                        title={t('settings')}
                        onClick={(e) => {
                            e.preventDefault()
                            history.push(settingsPath)
                        }}
                    >
                        <div
                            style={{
                                paddingLeft: 24,
                                marginRight: 12,
                                display: 'flex',
                                alignItems: 'center',
                            }}
                        >
                            <AiOutlineSetting />
                        </div>
                        <div
                            style={{
                                display: ctx.expanded ? 'block' : 'none',
                                fontSize: 14,
                            }}
                        >
                            {t('settings')}
                        </div>
                    </div>
                ) : (
                    <div
                        style={{
                            flexGrow: 1,
                            width: ctx.expanded ? sidebarExpandedWidth - sidebarFoldedWidth : sidebarFoldedWidth,
                        }}
                    />
                )}
                <div
                    role='button'
                    tabIndex={0}
                    onClick={handleExpandedClick}
                    style={{
                        cursor: 'pointer',
                        display: 'flex',
                        flexDirection: 'row',
                        height: 48,
                        alignItems: 'center',
                        background: ctx.expanded && isSettingsPage ? settingNavActiveBackground : 'transparent',
                    }}
                >
                    <div
                        style={{
                            display: 'inline-flex',
                            float: 'right',
                            alignSelf: 'center',
                            width: sidebarFoldedWidth,
                            justifyContent: 'center',
                        }}
                    >
                        {ctx.expanded ? <AiOutlineDoubleLeft /> : <AiOutlineDoubleRight />}
                    </div>
                </div>
            </div>
        </div>
    )
}
