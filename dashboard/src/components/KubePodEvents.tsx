import { formatMoment } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import { IWsRespSchema } from '@/schemas/websocket'
import { getEventTime, IKubeEventSchema } from '@/schemas/kube_event'
import qs from 'qs'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { toaster } from 'baseui/toast'
import { useOrganization } from '@/hooks/useOrganization'
import { Terminal as XtermTerminal } from 'xterm'
import { WebLinksAddon } from 'xterm-addon-web-links'
import { FitAddon } from 'xterm-addon-fit'
import { Tomorrow, Tomorrow_Night } from 'xterm-theme'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import _ from 'lodash'
import { SearchAddon } from 'xterm-addon-search'
import { Input } from 'baseui/input'
import { colors } from 'baseui/tokens'

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
    const elRef = useRef<null | HTMLDivElement>(null)
    const fitRef = useRef<null | FitAddon>(null)

    useEffect(() => {
        if (fitRef.current) {
            fitRef.current.fit()
        }
    }, [width, height])

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

    const themeType = useCurrentThemeType()
    const searchAddonRef = useRef<null | SearchAddon>(null)

    useEffect(() => {
        if (!open || !elRef.current) {
            return undefined
        }

        const terminal = new XtermTerminal({
            theme: themeType === 'light' ? Tomorrow : Tomorrow_Night,
            fontFamily: "Consolas, Menlo, 'Bitstream Vera Sans Mono', monospace, 'Powerline Symbols'",
            fontSize: 13,
            lineHeight: 1.2,
            macOptionIsMeta: true,
            cursorWidth: 1,
        })

        const searchAddon = new SearchAddon()
        const fitAddon = new FitAddon()
        terminal.loadAddon(new WebLinksAddon())
        terminal.loadAddon(fitAddon)
        terminal.loadAddon(searchAddon)
        searchAddonRef.current = searchAddon
        terminal.open(elRef.current)
        fitAddon.fit()
        fitRef.current = fitAddon

        const resizeHandler = () => {
            fitAddon.fit()
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
        let scrolling = false
        terminal.onScroll(() => {
            scrolling = true
        })
        const renderEvents = _.throttle((events: IKubeEventSchema[]) => {
            if (events.length === 0) {
                terminal.clear()
                terminal.writeln(t('no event'))
                return
            }
            events.forEach((event) => {
                const eventTime = getEventTime(event)
                const eventTimeStr = eventTime ? formatMoment(eventTime) : '-'
                let line
                if (podName) {
                    line = `[${eventTimeStr}] [${event.reason}] ${event.message}`
                } else {
                    line = `[${eventTimeStr}] [${event.involvedObject?.kind ?? '-'}] [${
                        event.involvedObject?.name ?? '-'
                    }] [${event.reason}] ${event.message}`
                }
                const line_ = line.toLowerCase()
                const hasError = line_.indexOf('error') !== -1 || line_.indexOf('fail') !== -1
                if (hasError) {
                    line = `\u001b[31m${line}\u001b[0m`
                }
                terminal.writeln(line)
            })
            if (!scrolling) {
                terminal.scrollToBottom()
            }
        }, 3000)
        let first = true
        let spinningTimer: number | undefined
        const spinner = '◰◳◲◱'.split('')
        let spinnerIdx = 0
        const spin = () => {
            if (selfClose) {
                return
            }
            if (first) {
                terminal.write(`\r ${t('loading...')} ${spinner[spinnerIdx]}`)
                spinnerIdx = (spinnerIdx + 1) % spinner.length
                spinningTimer = window.setTimeout(spin, 100)
            }
        }
        const stopSpin = () => {
            if (spinningTimer) {
                window.clearTimeout(spinningTimer)
                spinningTimer = undefined
            }
        }
        spin()
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
                if (first) {
                    first = false
                    stopSpin()
                    terminal.reset()
                }
                const events = resp.payload
                renderEvents(events)
            }
            const heartbeat = () => {
                if (ws?.readyState === ws?.OPEN) {
                    ws?.send('')
                }
                wsHeartbeatTimer = window.setTimeout(heartbeat, 20000)
            }
            ws.onopen = () => {
                heartbeat()
                resizeHandler()
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
            ws.onerror = (ev) => {
                // eslint-disable-next-line no-console
                console.log('onerror', ev)
                cancelHeartbeat()
            }
        }
        connect()
        window.addEventListener('resize', resizeHandler)
        return () => {
            stopSpin()
            searchAddonRef.current = null
            window.removeEventListener('resize', resizeHandler)
            terminal.dispose()
            cancelHeartbeat()
            selfClose = true
            ws?.close()
        }
    }, [wsUrl, open, organization?.name, t, podName, themeType])

    const searchOption = useMemo(
        () => ({
            decorations: {
                matchBackground: colors.yellow200,
                activeMatchBackground: colors.blue200,
                matchOverviewRuler: colors.yellow200,
                activeMatchColorOverviewRuler: colors.yellow200,
            },
        }),
        []
    )

    const onSearchValue_ = useCallback(
        (value: string) => {
            const searchAddon = searchAddonRef.current
            if (!searchAddon) {
                return
            }
            searchAddon.findNext(value, searchOption)
        },
        [searchOption]
    )

    const onSearchValue = useMemo(() => _.debounce(onSearchValue_, 300), [onSearchValue_])

    const onSearchKeyUp = useCallback(
        (e) => {
            if (e.key !== 'Enter') {
                return
            }
            const searchAddon = searchAddonRef.current
            if (!searchAddon) {
                return
            }
            if (e.shiftKey) {
                searchAddon.findPrevious(e.target.value, searchOption)
            } else {
                searchAddon.findNext(e.target.value, searchOption)
            }
        },
        [searchOption]
    )

    const [searchValue, setSearchValue] = useState('')

    useEffect(() => {
        onSearchValue(searchValue)
    }, [searchValue, onSearchValue])

    const onSearchChange = useCallback((e) => {
        setSearchValue(e.target.value)
    }, [])

    return (
        <div>
            <div
                style={{
                    display: 'flex',
                    flexDirection: 'row',
                    marginBottom: 10,
                }}
            >
                <div style={{ flexGrow: 1 }} />
                <Input
                    overrides={{
                        Root: {
                            style: {
                                width: '200px',
                                flexShrink: 0,
                            },
                        },
                    }}
                    size='mini'
                    clearable
                    value={searchValue}
                    onKeyUp={onSearchKeyUp}
                    onChange={onSearchChange}
                    placeholder={t('search')}
                />
            </div>
            <div style={{ height, width, overflow: 'hidden' }}>
                <div
                    ref={elRef}
                    style={{
                        flexGrow: 1,
                        width: '100%',
                        height: '100%',
                    }}
                />
            </div>
        </div>
    )
}
