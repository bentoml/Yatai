import React, { useRef, useEffect, useCallback } from 'react'
import { Terminal as XtermTerminal } from 'xterm'
import { WebLinksAddon } from 'xterm-addon-web-links'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'
import qs from 'qs'
import { toaster } from 'baseui/toast'
import { decode } from 'js-base64'
import { IKubePodSchema } from '@/schemas/kube_pod'
import { IWsRespSchema } from '@/schemas/websocket'
import { useOrganization } from '@/hooks/useOrganization'
import { Button } from 'baseui/button'
import { ProgressBar } from 'baseui/progress-bar'
import useTranslation from '@/hooks/useTranslation'
import { ImFinder } from 'react-icons/im'
import Upload from 'rc-upload'
import { GrStatusGood } from 'react-icons/gr'
import { useStyletron } from 'baseui'
import { Tab, Tabs } from 'baseui/tabs-motion'
import { AiOutlineCloudDownload, AiOutlineCloudUpload } from 'react-icons/ai'
import { Input } from 'baseui/input'
import Label from './Label'

interface ITerminalProps {
    clusterName: string
    deploymentName?: string
    namespace?: string
    podName: string
    containerName: string
    open?: boolean
    debug?: boolean
    fork?: boolean
    onGetGeneratedPod?: (pod: IKubePodSchema) => void
}

export default function Terminal({
    clusterName,
    deploymentName,
    namespace,
    podName,
    containerName: targetContainerName,
    open,
    debug,
    fork,
    onGetGeneratedPod,
}: ITerminalProps) {
    const elRef = useRef<null | HTMLDivElement>(null)
    const wsRef = useRef<null | WebSocket>(null)
    const fitRef = useRef<null | FitAddon>(null)
    const { organization } = useOrganization()
    const [isOpenFileManagerDrawer, setIsOpenFileManagerDrawer] = React.useState(false)
    const [t] = useTranslation()
    const [, theme] = useStyletron()
    const [fileManagerTabActiveKey, setFileManagerTabActiveKey] = React.useState('0')
    const [downloadPath, setDownloadPath] = React.useState('')
    const [containerName, setContainerName] = React.useState(debug ? undefined : targetContainerName)

    const [uploadingFiles, setUploadingFiles] = React.useState(
        [] as { name: string; percent: number; finished: boolean }[]
    )

    const beforeUpload = useCallback(
        (file) => {
            // 1G
            if (file.size > 10 * 1024 * 1024 * 1024) {
                toaster.negative(t('file size cannot exceed', ['10G']), { autoHideDuration: 5000 })
                return false
            }
            setUploadingFiles((files) => [
                {
                    name: file.name,
                    percent: 0,
                    finished: false,
                },
                ...files,
            ])
            return true
        },
        [t]
    )

    const handleUploadProgress = useCallback((progress, file) => {
        setUploadingFiles((files) =>
            files.map((f) => {
                if (f.name === file.name) {
                    return {
                        ...f,
                        percent: progress.percent,
                    }
                }
                return f
            })
        )
    }, [])

    const handleUploadSuccess = useCallback(
        (resp) => {
            const pieces = resp.dest_path.split('/')
            const fileName = pieces[pieces.length - 1]
            toaster.positive(t('upload sth successfully', [resp.dest_path]), { autoHideDuration: 5000 })
            setUploadingFiles((files) =>
                files.map((f) => {
                    if (f.name === fileName) {
                        return {
                            ...f,
                            finished: true,
                        }
                    }
                    return f
                })
            )
        },
        [t]
    )

    const handleUploadError = useCallback(
        (err) => {
            toaster.negative(`${t('upload file failed')}: ${err}`, { autoHideDuration: 5000 })
        },
        [t]
    )

    useEffect(() => {
        fitRef.current?.fit()
    }, [isOpenFileManagerDrawer])

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
            const msg = { type: 'resize', rows: terminal.rows, cols: terminal.cols }
            ws?.send(JSON.stringify(msg))
        }

        const wsUrl = deploymentName
            ? `${window.location.protocol === 'http:' ? 'ws:' : 'wss:'}//${
                  window.location.host
              }/ws/v1/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}/terminal?${qs.stringify(
                  {
                      organization_name: organization?.name,
                      pod_name: podName,
                      container_name: targetContainerName,
                      debug: debug ? 1 : 0,
                      fork: fork ? 1 : 0,
                  }
              )}`
            : `${window.location.protocol === 'http:' ? 'ws:' : 'wss:'}//${
                  window.location.host
              }/ws/v1/clusters/${clusterName}/terminal?${qs.stringify({
                  organization_name: organization?.name,
                  namespace,
                  pod_name: podName,
                  container_name: targetContainerName,
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
        ws.onclose = (ev) => {
            // eslint-disable-next-line no-console
            console.log('onclose', ev)
            terminal.write('\n!!! websocket closed !!!\n')
        }
        ws.onmessage = (event) => {
            try {
                const jsn = JSON.parse(event.data)
                const resp = jsn as IWsRespSchema<{ containerName: string }>
                if (resp.message) {
                    if (resp.type === 'error') {
                        toaster.negative(resp.message, {})
                    } else {
                        toaster.info(resp.message, {})
                    }
                }
                if (resp.payload?.containerName) {
                    setContainerName(resp.payload.containerName)
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
        ws.onerror = (ev) => {
            // eslint-disable-next-line no-console
            console.log('onerror', ev)
        }

        terminal.onData((input) => {
            const msg = { type: 'input', input }
            ws?.send(JSON.stringify(msg))
        })

        terminal.onResize(() => {
            resizeHandler()
        })

        window.addEventListener('resize', resizeHandler)

        // eslint-disable-next-line no-console
        console.log('terminal mounted')

        return () => {
            // eslint-disable-next-line no-console
            console.log('terminal unmount')
            fitRef.current = null
            terminal.dispose()
            window.removeEventListener('resize', resizeHandler)
            ws?.close()
        }
    }, [
        clusterName,
        targetContainerName,
        debug,
        deploymentName,
        fork,
        namespace,
        onGetGeneratedPod,
        open,
        organization?.name,
        podName,
    ])

    return (
        <div
            style={{
                display: 'flex',
                flexDirection: 'column',
                flexGrow: 1,
                width: '100%',
                height: '100%',
                gap: 10,
            }}
        >
            <div>
                <Button
                    onClick={() => setIsOpenFileManagerDrawer((isOpen) => !isOpen)}
                    startEnhancer={() => <ImFinder size={12} />}
                    size='mini'
                >
                    {isOpenFileManagerDrawer ? t('hide file manager') : t('show file manager')}
                </Button>
            </div>
            <div
                style={{
                    flexGrow: 1,
                    width: '100%',
                    height: '100%',
                    display: 'flex',
                    flexDirection: 'row',
                    gap: 20,
                }}
            >
                <div
                    ref={elRef}
                    style={{
                        flexGrow: 1,
                        width: isOpenFileManagerDrawer ? 'calc(100% - 600px)' : '100%',
                        height: '100%',
                    }}
                />
                {isOpenFileManagerDrawer && (
                    <div
                        style={{
                            width: 400,
                            flexShrink: 0,
                        }}
                    >
                        <Tabs
                            activeKey={fileManagerTabActiveKey}
                            onChange={({ activeKey }) => {
                                setFileManagerTabActiveKey(activeKey as string)
                            }}
                            activateOnFocus
                        >
                            <Tab title={t('upload file')} artwork={() => <AiOutlineCloudUpload />}>
                                <Upload
                                    beforeUpload={beforeUpload}
                                    onError={handleUploadError}
                                    onSuccess={handleUploadSuccess}
                                    onProgress={handleUploadProgress}
                                    action={`/api/v1/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}/pods/${podName}/containers/${containerName}/upload_file`}
                                >
                                    <div
                                        style={{
                                            width: '100%',
                                            padding: '40px 0',
                                            margin: '20px 0',
                                            textAlign: 'center',
                                            cursor: 'pointer',
                                            border: '2px dashed #eee',
                                        }}
                                    >
                                        <div style={{ fontWeight: 700, fontSize: '15px' }}>
                                            {t('click or drop file to this section')}
                                        </div>
                                        <span>{t('upload file to pod tips')}</span>
                                    </div>
                                </Upload>
                                {uploadingFiles.map((f, idx) => (
                                    <div key={uploadingFiles.length - idx} style={{ marginBottom: 6 }}>
                                        <div
                                            style={{
                                                display: 'flex',
                                                flexDirection: 'row',
                                                alignItems: 'center',
                                                gap: 8,
                                            }}
                                        >
                                            {f.finished && (
                                                <GrStatusGood
                                                    style={{
                                                        fill: theme.colors.positive300,
                                                    }}
                                                    size={12}
                                                />
                                            )}
                                            <div>{f.name}</div>
                                        </div>
                                        <ProgressBar
                                            overrides={{
                                                Root: {
                                                    style: {
                                                        display: f.finished ? 'none' : 'block',
                                                    },
                                                },
                                            }}
                                            value={f.percent}
                                        />
                                    </div>
                                ))}
                            </Tab>
                            <Tab title={t('download file')} artwork={() => <AiOutlineCloudDownload />}>
                                <div
                                    style={{
                                        display: 'flex',
                                        flexDirection: 'column',
                                        gap: 10,
                                    }}
                                >
                                    <div>
                                        <Label>{t('file path')}</Label>
                                        <Input
                                            placeholder={t('file path')}
                                            value={downloadPath}
                                            size='compact'
                                            onChange={(e) => {
                                                setDownloadPath((e.target as HTMLInputElement).value)
                                            }}
                                        />
                                    </div>
                                    <span>{t('please input the absolute path in the container')}</span>
                                    <Button
                                        disabled={!downloadPath.trim()}
                                        startEnhancer={() => <AiOutlineCloudDownload />}
                                        size='mini'
                                        onClick={() => {
                                            window.open(
                                                `/api/v1/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}/pods/${podName}/containers/${containerName}/download_file${qs.stringify(
                                                    {
                                                        path: downloadPath.trim(),
                                                    },
                                                    { addQueryPrefix: true }
                                                )}`
                                            )
                                        }}
                                    >
                                        {t('download')}
                                    </Button>
                                </div>
                            </Tab>
                        </Tabs>
                    </div>
                )}
            </div>
        </div>
    )
}
