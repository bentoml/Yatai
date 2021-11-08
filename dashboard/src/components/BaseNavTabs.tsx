import React, { useMemo } from 'react'
import { Tabs, Tab } from 'baseui/tabs-motion'
import { useHistory, useLocation } from 'react-router-dom'
import _ from 'lodash'
import { StatefulTooltip } from 'baseui/tooltip'
import { AiOutlineQuestionCircle } from 'react-icons/ai'
import { INavItem } from './BaseSidebar'

export interface IComposedNavTabsProps {
    style?: React.CSSProperties
    navStyle?: React.CSSProperties
}

export interface IBaseNavTabsProps extends IComposedNavTabsProps {
    navItems: INavItem[]
}

export function BaseNavTabs({ navItems }: IBaseNavTabsProps) {
    const history = useHistory()
    const location = useLocation()

    const activeItemId = useMemo(() => {
        const item = navItems
            .slice()
            .reverse()
            .find((item_) => _.startsWith(location.pathname, item_.path))
        return item?.path
    }, [location.pathname, navItems])

    return (
        <Tabs
            activeKey={activeItemId}
            onChange={({ activeKey }) => {
                history.push(activeKey as string)
            }}
            fill='fixed'
            activateOnFocus
        >
            {navItems.map((item) => {
                const Icon = item.icon
                return (
                    <Tab
                        overrides={{
                            TabPanel: {
                                style: {
                                    padding: '0px !important',
                                },
                            },
                        }}
                        disabled={item.disabled}
                        key={item.path}
                        title={
                            <div
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: 12,
                                    lineHeight: '24px',
                                    height: 24,
                                    textOverflow: 'ellipsis',
                                    whiteSpace: 'nowrap',
                                    overflow: 'hidden',
                                }}
                            >
                                {Icon && <Icon size={12} />}
                                <div
                                    style={{
                                        display: 'flex',
                                        alignItems: 'center',
                                        gap: 6,
                                    }}
                                >
                                    <span>{item.title}</span>
                                    {item.helpMessage && (
                                        <StatefulTooltip content={item.helpMessage} showArrow>
                                            <div
                                                style={{
                                                    display: 'inline-flex',
                                                    cursor: 'pointer',
                                                }}
                                            >
                                                <AiOutlineQuestionCircle />
                                            </div>
                                        </StatefulTooltip>
                                    )}
                                </div>
                            </div>
                        }
                    />
                )
            })}
        </Tabs>
    )
}
