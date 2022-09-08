import React, { useEffect, useRef, useCallback, useState } from 'react'
import qs from 'qs'
import useTranslation from '@/hooks/useTranslation'
import { v4 as uuidv4 } from 'uuid'
import { IWsReqSchema, IWsRespSchema } from '@/schemas/websocket'
import { toaster } from 'baseui/toast'
import { Select } from 'baseui/select'
import { useOrganization } from '@/hooks/useOrganization'
import { Terminal as XtermTerminal } from 'xterm'
import { WebLinksAddon } from 'xterm-addon-web-links'
import { FitAddon } from 'xterm-addon-fit'
import { Tomorrow, Tomorrow_Night } from 'xterm-theme'
import 'xterm/css/xterm.css'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import Card from './Card'
import Toggle from './Toggle'

interface ITailRequest {
    id: string
    tail_lines?: number
    container_name?: string
    follow: boolean
}

interface ILogProps {
    clusterName: string
    deploymentName?: string
    namespace: string
    podName: string
    open?: boolean
    width?: number | 'auto'
    height?: number | string
}

export default function Log({
    clusterName,
    deploymentName,
    namespace,
    podName,
    open,
    width = 300,
    height = 300,
}: ILogProps) {
    const elRef = useRef<null | HTMLDivElement>(null)
    const fitRef = useRef<null | FitAddon>(null)

    useEffect(() => {
        if (fitRef.current) {
            fitRef.current.fit()
        }
    }, [width, height])

    const [scroll, setScroll] = useState(true)

    const [t] = useTranslation()

    const { organization } = useOrganization()

    const wsUrl = deploymentName
        ? `${window.location.protocol === 'http:' ? 'ws:' : 'wss:'}//${
              window.location.host
          }/ws/v1/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}/tail?${qs.stringify({
              pod_name: podName,
              organization_name: organization?.name,
          })}`
        : `${window.location.protocol === 'http:' ? 'ws:' : 'wss:'}//${
              window.location.host
          }/ws/v1/clusters/${clusterName}/tail?${qs.stringify({
              namespace,
              pod_name: podName,
              organization_name: organization?.name,
          })}`

    const [tailLines, setTailLines] = useState(50)
    const [follow, setFollow] = useState(true)

    const reqIdRef = useRef('')
    const wsRef = useRef(null as null | WebSocket)
    const wsOpenRef = useRef(false)

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

    const themeType = useCurrentThemeType()

    useEffect(() => {
        if (!open || !elRef.current) {
            return undefined
        }

        const terminal = new XtermTerminal({
            theme: themeType === 'light' ? Tomorrow : Tomorrow_Night,
            fontFamily: "Consolas, Menlo, 'Bitstream Vera Sans Mono', monospace, 'Powerline Symbols'",
            fontSize: 13,
            macOptionIsMeta: true,
        })

        const fitAddon = new FitAddon()
        terminal.loadAddon(new WebLinksAddon())
        terminal.loadAddon(fitAddon)
        terminal.open(elRef.current)
        fitAddon.fit()
        fitRef.current = fitAddon
        terminal.focus()

        const resizeHandler = () => {
            fitAddon.fit()
        }

        let ws: WebSocket | undefined
        let selfClose = false
        const connect = () => {
            if (selfClose) {
                return
            }
            ws = new WebSocket(wsUrl)
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
                if (payload.type !== 'append') {
                    terminal.reset()
                }
                payload.items.forEach((item) => {
                    terminal.write(`${item}\r\n`)
                })
            }
            ws.onopen = () => {
                wsOpenRef.current = true
                if (ws) {
                    wsRef.current = ws
                }
                sendTailReq()
                resizeHandler()
            }
            ws.onclose = (ev) => {
                // eslint-disable-next-line no-console
                console.log('onclose', ev)
                wsOpenRef.current = false
                if (selfClose) {
                    return
                }
                setTimeout(connect, 3000)
            }
            ws.onerror = (ev) => {
                // eslint-disable-next-line no-console
                console.log('onerror', ev)
            }
        }
        connect()
        window.addEventListener('resize', resizeHandler)
        return () => {
            // eslint-disable-next-line no-console
            console.log('ws self close')
            window.removeEventListener('resize', resizeHandler)
            selfClose = true
            ws?.close(1000)
            wsRef.current = null
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [wsUrl, open, themeType])

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
                <div
                    ref={elRef}
                    style={{
                        flexGrow: 1,
                        width: '100%',
                        height: '100%',
                    }}
                />
            </div>
        </Card>
    )
}
