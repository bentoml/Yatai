import { useEffect, useRef, useCallback, useState, useMemo } from 'react'
import qs from 'qs'
import useTranslation from '@/hooks/useTranslation'
import { v4 as uuidv4 } from 'uuid'
import { IWsReqSchema, IWsRespSchema } from '@/schemas/websocket'
import { toaster } from 'baseui/toast'
import { Select, Option } from 'baseui/select'
import { useOrganization } from '@/hooks/useOrganization'
import { Terminal as XtermTerminal } from 'xterm'
import { WebLinksAddon } from 'xterm-addon-web-links'
import { FitAddon } from 'xterm-addon-fit'
import { Tomorrow, Tomorrow_Night } from 'xterm-theme'
import 'xterm/css/xterm.css'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import { colors } from 'baseui/tokens'
import _ from 'lodash'
import { Input } from 'baseui/input'
import { SearchAddon } from 'xterm-addon-search'
import { IContainerStatus, IKubePodSchema } from '@/schemas/kube_pod'
import { TbBrandDocker } from 'react-icons/tb'
import { IoMdList } from 'react-icons/io'
import { FaScroll } from 'react-icons/fa'
import { RiTimer2Line } from 'react-icons/ri'
import { ImListNumbered } from 'react-icons/im'
import Toggle from './Toggle'
import Card from './Card'

interface ITailRequest {
    id: string
    tail_lines?: number
    container_name?: string
    follow: boolean
}

interface ILogProps {
    clusterName: string
    deploymentName?: string
    pod: IKubePodSchema
    open?: boolean
    width?: number | 'auto'
    height?: number | string
}

export default function Log({ clusterName, deploymentName, pod, open, width = 300, height = 300 }: ILogProps) {
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
          }/ws/v1/clusters/${clusterName}/namespaces/${pod.namespace}/deployments/${deploymentName}/tail?${qs.stringify(
              {
                  pod_name: pod.name,
                  organization_name: organization?.name,
              }
          )}`
        : `${window.location.protocol === 'http:' ? 'ws:' : 'wss:'}//${
              window.location.host
          }/ws/v1/clusters/${clusterName}/tail?${qs.stringify({
              namespace: pod.namespace,
              pod_name: pod.name,
              organization_name: organization?.name,
          })}`

    const containerOptions = useMemo(() => {
        const options: Option[] = []
        const appendOption = (item: IContainerStatus, prefix?: string) => {
            options.push({
                id: item.name,
                label: `${prefix ? `[${prefix}] ` : ''}${
                    item.state.waiting !== undefined ? `${item.name} (${item.state.waiting.reason})` : item.name
                }`,
                disabled: item.state.waiting !== undefined,
            })
        }
        pod.raw_status?.initContainerStatuses?.forEach((item) => appendOption(item, 'init'))
        pod.raw_status?.containerStatuses?.forEach((item) => appendOption(item))
        return options
    }, [pod.raw_status?.containerStatuses, pod.raw_status?.initContainerStatuses])

    const [container, setContainer] = useState<string | undefined>(() => {
        const enabledOptions = containerOptions.filter((item) => !item.disabled)
        if (enabledOptions.length > 0) {
            return enabledOptions[enabledOptions.length - 1].id as string
        }
        return undefined
    })

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
                    container_name: container,
                    follow,
                },
            } as IWsReqSchema<ITailRequest>)
        )
    }, [container, follow, tailLines])

    useEffect(() => {
        sendTailReq()
    }, [sendTailReq])

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
                if (first) {
                    first = false
                    stopSpin()
                    terminal.reset()
                }
                payload.items.forEach((item) => {
                    terminal.writeln(item)
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
            stopSpin()
            searchAddonRef.current = null
            terminal.dispose()
            window.removeEventListener('resize', resizeHandler)
            selfClose = true
            ws?.close(1000)
            wsRef.current = null
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [wsUrl, open, themeType])

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
        <Card
            title={
                <div
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 4,
                    }}
                >
                    <IoMdList />
                    {t('view log')}
                </div>
            }
            extra={
                <div
                    style={{
                        display: 'flex',
                        flexDirection: 'row',
                        alignItems: 'center',
                        gap: 4,
                    }}
                >
                    <div
                        style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 4,
                        }}
                    >
                        <TbBrandDocker size={14} />
                        <div>Container</div>
                    </div>
                    <Select
                        overrides={{
                            Root: {
                                style: {
                                    minWidth: '220px',
                                },
                            },
                        }}
                        size='mini'
                        clearable={false}
                        searchable={false}
                        options={containerOptions}
                        value={[{ id: container }]}
                        onChange={(v) => {
                            setContainer(v.option?.id as string)
                        }}
                    />
                    <div
                        style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 4,
                            marginLeft: 12,
                        }}
                    >
                        <FaScroll size={14} />
                        <div>{t('scroll')}</div>
                    </div>
                    <Toggle value={scroll} onChange={setScroll} />
                    <div
                        style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 4,
                            marginLeft: 12,
                        }}
                    >
                        <RiTimer2Line size={14} />
                        <div>{t('realtime')}</div>
                    </div>
                    <Toggle value={follow} onChange={setFollow} />
                    <div
                        style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 4,
                            marginLeft: 12,
                        }}
                    >
                        <ImListNumbered size={14} />
                        <div>{t('lines')}</div>
                    </div>
                    <Select
                        size='mini'
                        overrides={{
                            Root: {
                                style: {
                                    minWidth: '80px',
                                },
                            },
                        }}
                        clearable={false}
                        searchable={false}
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
        </Card>
    )
}
