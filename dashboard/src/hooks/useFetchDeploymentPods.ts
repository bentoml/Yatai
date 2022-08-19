import { IWsRespSchema } from '@/schemas/websocket'
import { IKubePodSchema } from '@/schemas/kube_pod'
import { useEffect } from 'react'
import { toaster } from 'baseui/toast'
import qs from 'qs'
import { useOrganization } from './useOrganization'

export function useFetchDeploymentPods({
    clusterName,
    kubeNamespace,
    deploymentName,
    setPods,
    setPodsLoading,
    getErr,
}: {
    clusterName: string
    kubeNamespace: string
    deploymentName: string
    setPods: (pods: IKubePodSchema[]) => void
    setPodsLoading: (v: boolean) => void
    getErr?: (v: string) => void
}) {
    const { organization } = useOrganization()
    const wsUrl = `${window.location.protocol === 'http:' ? 'ws:' : 'wss:'}//${
        window.location.host
    }/ws/v1/clusters/${clusterName}/namespaces/${kubeNamespace}/deployments/${deploymentName}/pods${qs.stringify(
        {
            organization_name: organization?.name,
        },
        {
            addQueryPrefix: true,
        }
    )}`

    useEffect(() => {
        let ws: undefined | WebSocket
        let selfClose = false
        let wsHeartbeatTimer: undefined | number
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
            setPodsLoading(true)
            ws = new WebSocket(wsUrl)
            const heartbeat = () => {
                if (ws?.readyState === ws?.OPEN) {
                    ws?.send('')
                }
                wsHeartbeatTimer = window.setTimeout(heartbeat, 20000)
            }
            ws.onopen = () => heartbeat()
            ws.onerror = (ev) => {
                // eslint-disable-next-line no-console
                console.log('onerror', ev)
                cancelHeartbeat()
            }
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
                const resp = JSON.parse(event.data) as IWsRespSchema<IKubePodSchema[]>
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
                setPods(payload)
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
