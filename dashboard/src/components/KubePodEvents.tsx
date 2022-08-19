import { ScrollFollow } from 'react-lazylog'
import { formatMoment } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import { IWsRespSchema } from '@/schemas/websocket'
import { getEventTime, IKubeEventSchema } from '@/schemas/kube_event'
import qs from 'qs'
import { useEffect, useState } from 'react'
import { toaster } from 'baseui/toast'
import { useOrganization } from '@/hooks/useOrganization'
import LazyLog from './LazyLog'

interface IKubePodEventsProps {
    clusterName: string
    deploymentName?: string
    namespace: string
    podName?: string
    open?: boolean
    width?: number | 'auto'
    height?: number | string
}

export default function KubePodEvents({
    clusterName,
    deploymentName,
    namespace,
    podName,
    open,
    width,
    height,
}: IKubePodEventsProps) {
    const { organization } = useOrganization()

    const wsUrl = deploymentName
        ? `${window.location.protocol === 'http:' ? 'ws:' : 'wss:'}//${
              window.location.host
          }/ws/v1/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}/kube_events${qs.stringify(
              {
                  pod_name: podName,
                  organization_name: organization?.name,
              },
              {
                  addQueryPrefix: true,
              }
          )}`
        : `${window.location.protocol === 'http:' ? 'ws:' : 'wss:'}//${
              window.location.host
          }/ws/v1/clusters/${clusterName}/kube_events${qs.stringify(
              {
                  namespace,
                  pod_name: podName,
                  organization_name: organization?.name,
              },
              {
                  addQueryPrefix: true,
              }
          )}`

    const [t] = useTranslation()

    const [items, setItems] = useState<string[]>([])

    useEffect(() => {
        if (!open) {
            return undefined
        }
        let ws: WebSocket | undefined
        let selfClose = false
        let wsHeartbeatTimer: number | undefined
        const cancelHeartbeat = () => {
            if (wsHeartbeatTimer) {
                window.clearTimeout(wsHeartbeatTimer)
            }
            wsHeartbeatTimer = undefined
        }
        const connect = () => {
            if (!organization?.name) {
                return
            }
            if (selfClose) {
                return
            }
            ws = new WebSocket(wsUrl)
            ws.onmessage = (e) => {
                const resp = JSON.parse(e.data) as IWsRespSchema<IKubeEventSchema[]>
                if (resp.type !== 'success') {
                    toaster.negative(resp.message, {})
                    return
                }
                const events = resp.payload
                if (events.length === 0) {
                    setItems([t('no event')])
                    return
                }
                setItems(
                    events.map((event) => {
                        const eventTime = getEventTime(event)
                        const eventTimeStr = eventTime ? formatMoment(eventTime) : '-'
                        if (podName) {
                            return `[${eventTimeStr}] [${event.reason}] ${event.message}`
                        }
                        return `[${eventTimeStr}] [${event.involvedObject?.kind ?? '-'}] [${
                            event.involvedObject?.name ?? '-'
                        }] [${event.reason}] ${event.message}`
                    })
                )
            }
            const heartbeat = () => {
                if (ws?.readyState === ws?.OPEN) {
                    ws?.send('')
                }
                wsHeartbeatTimer = window.setTimeout(heartbeat, 20000)
            }
            ws.onopen = () => heartbeat()
            ws.onclose = (ev) => {
                // eslint-disable-next-line no-console
                console.log('onclose', ev)
                cancelHeartbeat()
                if (selfClose) {
                    return
                }
                setTimeout(connect, 3000)
            }
            ws.onerror = (ev) => {
                // eslint-disable-next-line no-console
                console.log('onerror', ev)
                cancelHeartbeat()
            }
        }
        connect()
        return () => {
            cancelHeartbeat()
            selfClose = true
            ws?.close()
        }
    }, [wsUrl, open, organization?.name, t, podName])

    return (
        <div style={{ height }}>
            <ScrollFollow
                startFollowing
                render={({ follow }) => (
                    <LazyLog
                        caseInsensitive
                        enableSearch
                        selectableLines
                        width={width}
                        text={items.length > 0 ? items.join('\n') : ' '}
                        follow={follow}
                    />
                )}
            />
        </div>
    )
}
