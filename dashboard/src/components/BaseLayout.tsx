import { headerHeight } from '@/consts'
import React from 'react'
import { Breadcrumbs } from 'baseui/breadcrumbs'
import { useHistory } from 'react-router-dom'
import { useStyletron } from 'baseui'
import { IComposedSidebarProps, INavItem } from './BaseSidebar'

export interface IBaseLayoutProps {
    children: React.ReactNode
    breadcrumbItems?: INavItem[]
    sidebar?: React.ComponentType<IComposedSidebarProps>
    contentStyle?: React.CSSProperties
    style?: React.CSSProperties
}

export default function BaseLayout({
    breadcrumbItems,
    children,
    sidebar: Sidebar,
    style,
    contentStyle,
}: IBaseLayoutProps) {
    const history = useHistory()
    const [, theme] = useStyletron()

    return (
        <main
            style={{
                height: '100vh',
                display: 'flex',
                flexFlow: 'row nowrap',
                justifyContent: 'space-between',
                ...style,
            }}
        >
            {Sidebar && <Sidebar style={{ marginTop: headerHeight }} />}
            <div
                style={{
                    overflowY: 'auto',
                    paddingTop: headerHeight,
                    flexGrow: 1,
                }}
            >
                <div
                    style={{
                        padding: '20px',
                        height: '100%',
                        boxSizing: 'border-box',
                        ...contentStyle,
                    }}
                >
                    {breadcrumbItems && (
                        <div style={{ marginBottom: 18 }}>
                            <Breadcrumbs
                                overrides={{
                                    List: {
                                        style: {
                                            display: 'flex',
                                            alignItems: 'center',
                                        },
                                    },
                                    ListItem: {
                                        style: {
                                            display: 'flex',
                                            alignItems: 'center',
                                        },
                                    },
                                }}
                            >
                                {breadcrumbItems.map((item, idx) => {
                                    const Icon = item.icon
                                    return (
                                        <div
                                            role='button'
                                            tabIndex={0}
                                            style={{
                                                fontSize: '13px',
                                                display: 'flex',
                                                alignItems: 'center',
                                                gap: 6,
                                                cursor: idx !== breadcrumbItems.length - 1 ? 'pointer' : undefined,
                                            }}
                                            key={item.path}
                                            onClick={
                                                idx !== breadcrumbItems.length - 1
                                                    ? () => {
                                                          history.push(item.path)
                                                      }
                                                    : undefined
                                            }
                                        >
                                            {Icon && <Icon size={12} />}
                                            <span
                                                style={{
                                                    borderBottomStyle: 'solid',
                                                    borderBottomWidth: 1,
                                                    borderBottomColor:
                                                        idx !== breadcrumbItems.length - 1
                                                            ? theme.colors.contentPrimary
                                                            : 'transparent',
                                                }}
                                            >
                                                {item.title}
                                            </span>
                                        </div>
                                    )
                                })}
                            </Breadcrumbs>
                        </div>
                    )}

                    {children}
                </div>
            </div>
        </main>
    )
}
