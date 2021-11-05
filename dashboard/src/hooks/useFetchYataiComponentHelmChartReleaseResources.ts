import { IWsRespSchema } from '@/schemas/websocket'
import { useEffect, useRef } from 'react'
import { IKubeResourceSchema } from '@/schemas/kube_resource'
import { YataiComponentType } from '@/schemas/yatai_component'
import { toaster } from 'baseui/toast'

export function useFetchYataiComponentHelmChartReleaseResources(
    clusterName: string,
    compType: YataiComponentType,
    setKubeResources: (pods: IKubeResourceSchema[]) => void,
    setKubeResourcesLoading: (v: boolean) => void
) {
    const wsUrl = `${window.location.protocol === 'http:' ? 'ws:' : 'wss:'}//${
        window.location.host
    }/ws/v1/clusters/${clusterName}/yatai_components/${compType}/helm_chart_release_resources`

    const wsRef = useRef(undefined as undefined | WebSocket)
    const wsHeartbeatTimerRef = useRef(undefined as undefined | number)

    useEffect(() => {
        let ws: undefined | WebSocket
        let selfClose = false
        const cancelHeartbeat = () => {
            if (wsHeartbeatTimerRef.current) {
                window.clearTimeout(wsHeartbeatTimerRef.current)
            }
            wsHeartbeatTimerRef.current = undefined
        }
        const connect = () => {
            if (wsRef.current) {
                wsRef.current.close()
            }
            setKubeResourcesLoading(true)
            ws = new WebSocket(wsUrl)
            const heartbeat = () => {
                if (ws?.readyState === ws?.OPEN) {
                    ws?.send('')
                }
                wsHeartbeatTimerRef.current = window.setTimeout(heartbeat, 20000)
            }
            wsRef.current = ws
            ws.onopen = () => heartbeat()
            // eslint-disable-next-line no-console
            ws.onerror = () => console.log('onerror')
            ws.onclose = () => {
                cancelHeartbeat()
                if (selfClose) {
                    return
                }
                setTimeout(connect, 3000)
            }
            ws.onmessage = (event) => {
                if (selfClose) {
                    return
                }
                const resp = JSON.parse(event.data) as IWsRespSchema<IKubeResourceSchema[]>
                if (resp.type === 'error') {
                    toaster.negative(resp.message, {})
                    selfClose = true
                    ws?.close()
                    return
                }
                const { payload } = resp
                setKubeResources(payload)
                setKubeResourcesLoading(false)
            }
        }
        connect()
        return () => {
            cancelHeartbeat()
            selfClose = true
            ws?.close()
        }
    }, [setKubeResources, setKubeResourcesLoading, wsUrl])
}
