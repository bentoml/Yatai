import React from 'react'
import { ScrollFollow, LazyLog } from 'react-lazylog'
import { formatTime } from '@/utils/datetime'
import { createUseStyles } from 'react-jss'
import { IThemedStyleProps } from '@/interfaces/IThemedStyle'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import useTranslation from '@/hooks/useTranslation'
import { useStyletron } from 'baseui'
import { IWsRespSchema } from '@/schemas/websocket'
import { IKubeEventSchema } from '@/schemas/kube_event'

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

interface IAppKubeEventsProps {
    orgName: string
    clusterName: string
    deploymentName: string
    open?: boolean
    width?: number | 'auto'
    height?: number | string
}

export default ({ orgName, clusterName, deploymentName, open, width, height }: IAppKubeEventsProps) => {
    const [, theme] = useStyletron()
    const themeType = useCurrentThemeType()
    const styles = useStyles({ theme, themeType })

    const wsUrl = `${window.location.protocol === 'http:' ? 'ws:' : 'wss:'}//${
        window.location.host
    }/ws/v1/orgs/${orgName}/clusters/${clusterName}/deployments/${deploymentName}/kube_events`

    let logContainerStyle: React.CSSProperties = {
        background: theme.colors.backgroundPrimary,
    }

    if (width !== 'auto') {
        logContainerStyle = {
            ...logContainerStyle,
            width,
            height: typeof height === 'number' ? height - 46 - 40 : height,
        }
    }

    const [t] = useTranslation()

    if (!open) {
        return <></>
    }

    return (
        <ScrollFollow
            startFollowing
            render={({ follow }) => (
                <LazyLog
                    height={200}
                    url={wsUrl}
                    lineClassName={styles.line}
                    style={logContainerStyle}
                    websocket
                    websocketOptions={{
                        formatMessage: (e) => {
                            const resp = JSON.parse(e) as IWsRespSchema<IKubeEventSchema | null>
                            const event = resp.payload
                            if (!event) {
                                return t('no event')
                            }
                            return `[${event.lastTimestamp ? formatTime(event.lastTimestamp) : '-'}] [${
                                event.involvedObject?.kind ?? '-'
                            }] [${event.involvedObject?.name ?? '-'}] [${event.reason}] ${event.message}`
                        },
                    }}
                    follow={follow}
                />
            )}
        />
    )
}
