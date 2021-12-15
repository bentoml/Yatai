import { toaster } from 'baseui/toast'
import { IWsReqSchema, IWsRespSchema } from '@/schemas/websocket'
/* eslint-disable @typescript-eslint/no-explicit-any */
import { ResourceType } from '@/schemas/resource'
import { ISubscriptionReqSchema, ISubscriptionRespSchema } from '@/schemas/subscription'
import { useCallback, useEffect, useRef } from 'react'
import _ from 'lodash'

type CB = (payload: any) => void

export interface ICBItem {
    resourceType: ResourceType
    resourceUids: string[]
    cb: CB
}

export function useSubscription() {
    const cbItemsRef = useRef<ICBItem[]>([])
    const wsRef = useRef<WebSocket | undefined>()
    const wsUrl = `${window.location.protocol === 'http:' ? 'ws:' : 'wss:'}//${
        window.location.host
    }/ws/v1/subscription/resource`

    const subscribe = useCallback((cbItem: ICBItem) => {
        if (!wsRef.current) {
            return
        }
        wsRef.current.send(
            JSON.stringify({
                type: 'data',
                payload: {
                    action: 'subscribe',
                    resource_type: cbItem.resourceType,
                    resource_uids: cbItem.resourceUids,
                },
            } as IWsReqSchema<ISubscriptionReqSchema>)
        )
        cbItemsRef.current.push(cbItem)
    }, [])

    const unsubscribe = useCallback((cbItem: ICBItem) => {
        if (!wsRef.current) {
            return
        }
        const oldUids = cbItemsRef.current.reduce((p, c) => {
            if (c.resourceType !== cbItem.resourceType) {
                return p
            }
            return [...p, ...c.resourceUids]
        }, [] as string[])
        cbItemsRef.current = cbItemsRef.current.filter((cbItem_) => !_.isEqual(cbItem_, cbItem))
        const newUids = cbItemsRef.current.reduce((p, c) => {
            if (c.resourceType !== cbItem.resourceType) {
                return p
            }
            return [...p, ...c.resourceUids]
        }, [] as string[])
        wsRef.current.send(
            JSON.stringify({
                type: 'data',
                payload: {
                    action: 'unsubscribe',
                    resource_type: cbItem.resourceType,
                    resource_uids: _.difference(oldUids, newUids),
                },
            } as IWsReqSchema<ISubscriptionReqSchema>)
        )
    }, [])

    const wsHeartbeatTimerRef = useRef<number | undefined>()

    useEffect(() => {
        let selfClose = false
        const cancelHeartbeat = () => {
            if (wsHeartbeatTimerRef.current) {
                window.clearTimeout(wsHeartbeatTimerRef.current)
            }
            wsHeartbeatTimerRef.current = undefined
        }
        let ws: undefined | WebSocket
        const connect = () => {
            if (wsRef.current) {
                wsRef.current.close()
                return
            }
            ws = new WebSocket(wsUrl)
            selfClose = false
            const heartbeat = () => {
                if (ws?.readyState === ws?.OPEN) {
                    ws?.send(
                        JSON.stringify({
                            type: 'heartbeat',
                        } as IWsReqSchema<undefined>)
                    )
                }
                wsHeartbeatTimerRef.current = window.setTimeout(heartbeat, 20000)
            }
            ws.onopen = () => {
                wsRef.current = ws
                heartbeat()
                cbItemsRef.current.forEach((cbItem) => {
                    ws?.send(
                        JSON.stringify({
                            type: 'data',
                            payload: {
                                action: 'subscribe',
                                resource_type: cbItem.resourceType,
                                resource_uids: cbItem.resourceUids,
                            },
                        } as IWsReqSchema<ISubscriptionReqSchema>)
                    )
                })
            }

            ws.onclose = () => {
                cancelHeartbeat()
                if (selfClose) {
                    return
                }
                setTimeout(connect, 3000)
            }

            ws.onmessage = (event) => {
                const resp = JSON.parse(event.data) as IWsRespSchema<ISubscriptionRespSchema<any>>
                if (resp.type === 'error') {
                    toaster.negative(resp.message, {})
                    return
                }
                const { payload } = resp
                cbItemsRef.current.forEach((cbItem) => {
                    if (cbItem.resourceType !== payload.resource_type) {
                        return
                    }
                    if (cbItem.resourceUids.indexOf(payload.payload.uid) < 0) {
                        return
                    }
                    cbItem.cb(payload.payload)
                })
            }
            ws.onerror = () => {
                // eslint-disable-next-line no-console
                console.log('onerror')
            }
        }
        connect()
        return () => {
            cancelHeartbeat()
            selfClose = true
            ws?.close()
        }
    }, [wsUrl])

    return {
        subscribe,
        unsubscribe,
    }
}
