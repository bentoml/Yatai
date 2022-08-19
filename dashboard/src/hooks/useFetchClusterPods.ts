import { IWsRespSchema } from '@/schemas/websocket'
import { IKubePodSchema } from '@/schemas/kube_pod'
import { useEffect, useRef } from 'react'
import { toaster } from 'baseui/toast'
import qs from 'qs'
import { useOrganization } from './useOrganization'

export function useFetchClusterPods({
    clusterName,
    namespace,
    selector,
    setPods,
    setPodsLoading,
    getErr,
}: {
    clusterName: string
    namespace: string
    selector: string
    setPods: (pods: IKubePodSchema[]) => void
    setPodsLoading: (v: boolean) => void
    getErr?: (v: string) => void
}) {
    const { organization } = useOrganization()
    const wsUrl = `${window.location.protocol === 'http:' ? 'ws:' : 'wss:'}//${
        window.location.host
    }/ws/v1/clusters/${clusterName}/pods${qs.stringify(
        {
            organization_name: organization?.name,
            namespace,
            selector,
        },
        {
            addQueryPrefix: true,
        }
    )}`

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
            if (!organization?.name) {
                return
            }
            if (selfClose) {
                return
            }
            setPodsLoading(true)
            ws = new WebSocket(wsUrl)
            const heartbeat = () => {
                if (ws?.readyState === ws?.OPEN) {
                    ws?.send('')
                }
                wsHeartbeatTimerRef.current = window.setTimeout(heartbeat, 20000)
            }
            ws.onopen = () => heartbeat()
            // eslint-disable-next-line no-console
            ws.onerror = (ev) => console.log('onerror', ev)
            ws.onclose = (ev) => {
                // eslint-disable-next-line no-console
                console.log('onclose', ev)
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
                const resp = JSON.parse(event.data) as IWsRespSchema<IKubePodSchema[] | null>
                if (resp.type === 'error') {
                    if (getErr) {
                        getErr(resp.message)
                    } else {
                        toaster.negative(resp.message, {})
                    }
                    selfClose = true
                    ws?.close()
                    return
                }
                const { payload } = resp
                setPods(payload ?? [])
                setPodsLoading(false)
            }
        }
        connect()
        return () => {
            cancelHeartbeat()
            selfClose = true
            ws?.close()
        }
    }, [getErr, organization?.name, setPods, setPodsLoading, wsUrl])
}
