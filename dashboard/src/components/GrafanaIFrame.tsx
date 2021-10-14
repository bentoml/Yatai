import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import { processUrl } from '@/utils'
import { useStyletron } from 'baseui'
import { StyledLink } from 'baseui/link'
import { Skeleton } from 'baseui/skeleton'
import { AiOutlineLink } from 'react-icons/ai'

import classNames from 'classnames'
import color from 'color'
import qs from 'qs'
import React, { SyntheticEvent, useCallback, useState } from 'react'
import { createUseStyles } from 'react-jss'

const useStyles = createUseStyles({
    wrapper: {
        'flex-grow': 1,
        'position': 'relative',
        '&:hover': {
            '& $link': {
                display: 'block',
            },
        },
    },
    link: {
        display: 'none',
        position: 'absolute',
        top: 4,
        right: 6,
    },
    iframe: {
        border: 0,
        width: '100%',
        height: '100%',
    },
})

interface IGrafanaIFrameProps {
    className?: string
    style?: React.CSSProperties
    title: string
    baseUrl: string
    pathname: string
    query: Record<string, unknown>
    externalPathname?: string
    externalQuery?: Record<string, string | number | boolean | undefined | null>
    iframeCSS?: string
    onLoad?: () => void
}

export default ({
    className,
    style,
    title,
    baseUrl: _baseUrl,
    pathname,
    query,
    externalPathname,
    externalQuery,
    iframeCSS,
    onLoad,
}: IGrafanaIFrameProps) => {
    const themeType = useCurrentThemeType()
    const [, theme] = useStyletron()
    const [show, setShow] = useState(false)
    const styles = useStyles()

    const handleIFrameLoad = useCallback(
        (e: SyntheticEvent<HTMLIFrameElement>) => {
            const target = e.target as HTMLIFrameElement | undefined
            target?.contentWindow?.postMessage(
                {
                    style: `
                    .explore-toolbar {
                        margin-top: 0px;
                    }
                    .css-ytckdi-exploreMain {
                        margin-top: 0px;
                    }
                    .sidemenu, .navbar-page-btn, .explore-toolbar-header-title, .explore-ds-picker, .css-kj45dn-queryContainer, .css-18dr9jf-queryContainer, .css-hz279r-collapse__header--collapsed, .css-1ugehg8-collapse__header--collapsed {
                        display: none !important;
                    }
                    .panel-container {
                        border: 0 !important;
                    }
                    .css-1baakqg,
                    .css-uwlvbv {
                        border-color: ${theme.borders.border100.borderColor} !important;
                    }
                    .css-122caa4-button {
                        background: ${theme.colors.backgroundLightAccent} !important;
                        color: ${theme.colors.contentAccent} !important;
                    }
                    .css-1i8hcrs,
                    .css-1767p2b,
                    .explore-wrapper, .panel-container, .explore, body, .main-view {
                        background-color: ${theme.colors.backgroundPrimary} !important;
                    }
                    ${
                        themeType === 'light'
                            ? `
                            .explore-container .panel-container {
                                background-color: ${color(theme.colors.backgroundPrimary)
                                    .darken(0.001)
                                    .rgb()
                                    .string()} !important;
                            }
                            .explore-wrapper .logs-panel-options {
                                background-color: ${color(theme.colors.backgroundPrimary)
                                    .darken(0.01)
                                    .rgb()
                                    .string()} !important;
                            }
                            main.css-1rxjq6w-logsMain {
                                background-color: ${color(theme.colors.backgroundPrimary)
                                    .darken(0.01)
                                    .rgb()
                                    .string()} !important;
                            }
                            `
                            : `
                    .react-calendar, .css-1lwgu2p, .css-iq0sfc, .dropdown-menu--menu, .dropdown-menu--navbar, .dropdown-menu--sidemenu, .css-ajr8sn, .gf-form-select-box__menu {
                        background-color: ${color(theme.colors.backgroundPrimary)
                            .darken(0.3)
                            .rgb()
                            .string()} !important;
                    }
                    .css-uwlvbv {
                        border-width: 0px !important;
                    }
                    button, input, .navbar-button {
                        background-color: ${color(theme.colors.backgroundPrimary)
                            .darken(0.1)
                            .rgb()
                            .string()} !important;
                    }
                    .explore-wrapper .panel-container {
                        border-color: ${color(theme.colors.backgroundPrimary).darken(0.2).rgb().string()} !important;
                    }
                    .explore-wrapper .logs-panel-options {
                        background-color: ${color(theme.colors.backgroundPrimary)
                            .darken(0.3)
                            .rgb()
                            .string()} !important;
                    }
                    .logs-panel tr:hover, .panel-header:hover {
                        background-color: ${color(theme.colors.backgroundPrimary)
                            .darken(0.2)
                            .rgb()
                            .string()} !important;
                    }
                    .css-cssveg {
                        background-color: ${color(theme.colors.backgroundPrimary)
                            .darken(0.2)
                            .rgb()
                            .string()} !important;
                    }
                    .css-thhc72-nameWrapper {
                        background-color: ${color(theme.colors.backgroundPrimary)
                            .darken(0.3)
                            .rgb()
                            .string()} !important;
                    }
                    .css-f2mhmw-SpanTreeOffsetParent:hover {
                        background-color: ${color(theme.colors.backgroundPrimary)
                            .darken(0.2)
                            .rgb()
                            .string()} !important;
                    }
                    .css-cj34uv-logs-row-hoverBackground {
                        background-color: ${color(theme.colors.backgroundPrimary)
                            .darken(0.2)
                            .rgb()
                            .string()} !important;
                    }
                    .css-1xl0vdh-logs-row-hoverBackground-logDetailsDefaultCursor {
                        background-color: ${color(theme.colors.backgroundPrimary)
                            .darken(0.2)
                            .rgb()
                            .string()} !important;
                    }
                    ${iframeCSS ? String(iframeCSS) : ''}
`
                    }
                    `,
                },
                '*'
            )
            setTimeout(() => {
                setShow(true)
                onLoad?.()
            }, 600)
        },
        [iframeCSS, onLoad, theme.colors.background, themeType]
    )

    const baseUrl = processUrl(_baseUrl)

    const url = `${baseUrl}${pathname}${qs.stringify(
        {
            theme: themeType,
            ...query,
        },
        { addQueryPrefix: true, strictNullHandling: true }
    )}`

    let externalUrl: string | undefined

    if (externalPathname) {
        externalUrl = `${baseUrl}${externalPathname}${qs.stringify(
            {
                theme: themeType,
                ...externalQuery,
            },
            { addQueryPrefix: true, strictNullHandling: true }
        )}`
    }

    // Use key to update iframe without affecting browser history
    return (
        <div className={classNames(styles.wrapper, className)} style={style}>
            <iframe
                key={url}
                className={styles.iframe}
                style={{
                    visibility: show ? 'visible' : 'hidden',
                }}
                onLoad={handleIFrameLoad}
                src={url}
                frameBorder='0'
                title={title}
            />
            {externalUrl && (
                <StyledLink className={styles.link} target='_blank' href={externalUrl}>
                    <AiOutlineLink />
                </StyledLink>
            )}
            <div
                style={{
                    position: 'absolute',
                    top: 0,
                    bottom: 0,
                    left: 0,
                    right: 0,
                    display: show ? 'none' : 'block',
                }}
            >
                <Skeleton animation rows={3} />
            </div>
        </div>
    )
}
