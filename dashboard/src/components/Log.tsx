import React, { useEffect, useRef, useCallback, useState } from 'react'
import qs from 'qs'
import { ScrollFollow } from 'react-lazylog'
import useTranslation from '@/hooks/useTranslation'
import { v4 as uuidv4 } from 'uuid'
import { IWsReqSchema, IWsRespSchema } from '@/schemas/websocket'
import { toaster } from 'baseui/toast'
import { Select } from 'baseui/select'
import Card from './Card'
import Toggle from './Toggle'
import LazyLog from './LazyLog'

interface ITailRequest {
    id: string
    tail_lines?: number
    container_name?: string
    follow: boolean
}

interface ILogProps {
    orgName: string
    clusterName: string
    deploymentName: string
    podName: string
    open?: boolean
    width?: number | 'auto'
    height?: number | string
}

export default ({ orgName, clusterName, deploymentName, podName, open, width = 300, height = 300 }: ILogProps) => {
    const [scroll, setScroll] = useState(true)

    const [t] = useTranslation()

    const wsUrl = `${window.location.protocol === 'http:' ? 'ws:' : 'wss:'}//${
        window.location.host
    }/ws/v1/orgs/${orgName}/clusters/${clusterName}/deployments/${deploymentName}/tail?${qs.stringify({
        pod_name: podName,
    })}`

    const [items, setItems] = useState<string[]>([])
    const [tailLines, setTailLines] = useState(50)
    const [follow, setFollow] = useState(true)

    const reqIdRef = useRef('')
    const wsRef = useRef(null as null | WebSocket)
    const wsOpenRef = useRef(false)
    const selfCloseRef = useRef(false)

    const sendTailReq = useCallback(() => {
        if (!wsOpenRef.current) {
            return
        }
        const id = uuidv4()
        reqIdRef.current = id
        wsRef.current?.send(
            JSON.stringify({
                type: 'data',
                payload: {
                    id,
                    tail_lines: tailLines,
                    follow,
                },
            } as IWsReqSchema<ITailRequest>)
        )
    }, [follow, tailLines])

    useEffect(() => {
        sendTailReq()
    }, [sendTailReq])

    useEffect(() => {
        if (!open) {
            return undefined
        }
        let ws: WebSocket | undefined
        const connect = () => {
            ws = new WebSocket(wsUrl)
            selfCloseRef.current = false
            ws.onmessage = (event) => {
                const resp = JSON.parse(event.data) as IWsRespSchema<{
                    req_id: string
                    type: 'append' | 'replace'
                    items: string[]
                }>
                if (resp.type !== 'success') {
                    toaster.negative(resp.message, {})
                    return
                }
                const { payload } = resp
                if (payload.req_id !== reqIdRef.current) {
                    return
                }
                if (payload.type === 'append') {
                    setItems((_items) => [..._items, ...payload.items])
                } else {
                    setItems(payload.items)
                }
            }
            ws.onopen = () => {
                wsOpenRef.current = true
                if (ws) {
                    wsRef.current = ws
                }
                sendTailReq()
            }
            ws.onclose = () => {
                wsOpenRef.current = false
                if (selfCloseRef.current) {
                    return
                }
                connect()
            }
        }
        connect()
        return () => {
            ws?.close()
            selfCloseRef.current = true
            wsRef.current = null
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [wsUrl, open])

    return (
        <Card
            title={t('view log')}
            extra={
                <div
                    style={{
                        display: 'flex',
                        flexDirection: 'row',
                        alignItems: 'center',
                        gap: 4,
                    }}
                >
                    <div>{t('scroll')}</div>
                    <Toggle value={scroll} onChange={setScroll} />
                    <div style={{ marginLeft: 12 }}>{t('realtime')}</div>
                    <Toggle value={follow} onChange={setFollow} />
                    <div style={{ marginLeft: 12 }}>{t('lines')}</div>
                    <Select
                        options={[
                            {
                                label: '50',
                                id: 50,
                            },
                            {
                                label: '100',
                                id: 100,
                            },
                            {
                                label: '200',
                                id: 200,
                            },
                            {
                                label: '1000',
                                id: 1000,
                            },
                            {
                                label: '5000',
                                id: 5000,
                            },
                            {
                                label: '10000',
                                id: 10000,
                            },
                        ]}
                        value={[{ id: tailLines }]}
                        onChange={(v) => {
                            setTailLines(v.option?.id as number)
                        }}
                    />
                </div>
            }
        >
            <div style={{ height }}>
                <ScrollFollow
                    startFollowing={scroll}
                    render={({ follow: follow_ }) => (
                        <LazyLog
                            caseInsensitive
                            width={width}
                            enableSearch
                            selectableLines
                            text={items.length > 0 ? items.join('\n') : ' '}
                            follow={follow_}
                        />
                    )}
                />
            </div>
        </Card>
    )
}
