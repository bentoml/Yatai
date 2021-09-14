import React, { useRef, useEffect } from 'react'
import { Terminal as XtermTerminal } from 'xterm'
import { WebLinksAddon } from 'xterm-addon-web-links'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'
import qs from 'qs'
import { decode } from 'js-base64'
import { IKubePodSchema } from '@/schemas/kube_pod'
import { IWsRespSchema } from '@/schemas/websocket'
import { toaster } from 'baseui/toast'

interface ITerminalProps {
    orgName?: string
    clusterName?: string
    deploymentName?: string
    podName: string
    containerName: string
    open?: boolean
    width?: number
    height?: number
    debug?: boolean
    fork?: boolean
    onGetGeneratedPod?: (pod: IKubePodSchema) => void
}

export default function Terminal({
    orgName,
    clusterName,
    deploymentName,
    podName,
    containerName,
    open,
    width,
    height,
    debug,
    fork,
    onGetGeneratedPod,
}: ITerminalProps) {
    const elRef = useRef<null | HTMLDivElement>(null)
    const wsRef = useRef<null | WebSocket>(null)
    const fitRef = useRef<null | FitAddon>(null)

    useEffect(() => {
        if (fitRef.current) {
            fitRef.current.fit()
        }
    }, [width, height])

    useEffect(() => {
        if (!open || !elRef.current) {
            return undefined
        }

        const terminal = new XtermTerminal({
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
        let ws = wsRef.current
        if (ws) {
            return undefined
        }

        const resizeHandler = () => {
            fitAddon.fit()
            const msg = { type: 'resize', rows: terminal.rows, cols: terminal.cols }
            ws?.send(JSON.stringify(msg))
        }

        const wsUrl = `${window.location.protocol === 'http:' ? 'ws:' : 'wss:'}//${
            window.location.host
        }/ws/v1/orgs/${orgName}/clusters/${clusterName}/deployments/${deploymentName}/terminal?${qs.stringify({
            pod_name: podName,
            container_name: containerName,
            debug: debug ? 1 : 0,
            fork: fork ? 1 : 0,
        })}`
        ws = new WebSocket(wsUrl)
        wsRef.current = ws
        ws.onopen = () => {
            // eslint-disable-next-line no-console
            console.log('onopen')
            resizeHandler()
        }
        ws.onclose = () => {
            // eslint-disable-next-line no-console
            console.log('onclose')
            terminal.write('\n!!! websocket closed !!!\n')
        }
        ws.onmessage = (event) => {
            try {
                const resp = JSON.parse(event.data) as IWsRespSchema<string>
                if (resp.message) {
                    if (resp.type === 'error') {
                        toaster.negative(resp.message, {})
                    } else {
                        toaster.info(resp.message, {})
                    }
                }
                return
            } catch {
                //
            }
            if (onGetGeneratedPod) {
                try {
                    const resp = JSON.parse(event.data)
                    if (resp.is_mcd_msg && resp.pod) {
                        onGetGeneratedPod(resp.pod)
                        return
                    }
                } catch {
                    //
                }
            }
            const data = decode(event.data)
            terminal.write(data)
        }
        ws.onerror = () => {
            // eslint-disable-next-line no-console
            console.log('onerror')
        }

        terminal.onData((input) => {
            const msg = { type: 'input', input }
            ws?.send(JSON.stringify(msg))
        })

        terminal.onResize(() => {
            resizeHandler()
        })

        window.addEventListener('resize', resizeHandler)

        return () => {
            window.removeEventListener('resize', resizeHandler)
            ws?.close()
        }
    }, [clusterName, containerName, debug, deploymentName, fork, onGetGeneratedPod, open, orgName, podName])

    return (
        <div
            ref={elRef}
            style={{
                flexGrow: 1,
                width: '100%',
                height: '100%',
            }}
        />
    )
}
