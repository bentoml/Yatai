import { useCluster } from '@/hooks/useCluster'
import { useDeployment } from '@/hooks/useDeployment'
import { useOrganization } from '@/hooks/useOrganization'
import useTranslation from '@/hooks/useTranslation'
import { IKubePodSchema } from '@/schemas/kube_pod'
import { formatDateTime } from '@/utils/datetime'
import { Modal, ModalBody, ModalHeader } from 'baseui/modal'
import { IoMdList } from 'react-icons/io'
import { GoTerminal } from 'react-icons/go'
import { MdEventNote } from 'react-icons/md'
import React, { useState, useCallback } from 'react'
import { StatefulTooltip } from 'baseui/tooltip'
import { Button } from 'baseui/button'
import { AiOutlineDashboard, AiOutlineQuestionCircle } from 'react-icons/ai'
import { useFetchYataiComponents } from '@/hooks/useFetchYataiComponents'
import {
    StyledTable,
    StyledTableBodyCell,
    StyledTableBodyRow,
    StyledTableHeadCell,
    StyledTableHeadRow,
} from 'baseui/table-semantic'
import { Skeleton } from 'baseui/skeleton'
import { resourceIconMapping } from '@/consts'
import Log from './Log'
import { PodStatus } from './PodStatuses'
import Terminal from './Terminal'
import KubePodEvents from './KubePodEvents'
import Toggle from './Toggle'
import Label from './Label'
import LokiLog from './LokiLog'
import PodMonitor from './PodMonitor'

export interface IPodListProps {
    loading?: boolean
    clusterName?: string
    groupByRunner?: boolean
    pods: IKubePodSchema[]
}

export default function PodList({
    loading = false,
    clusterName: clusterName_,
    pods,
    groupByRunner = false,
}: IPodListProps) {
    const [t] = useTranslation()
    const { organization } = useOrganization()
    const { cluster } = useCluster()
    const { deployment } = useDeployment()
    const [desiredShowLogsPod, setDesiredShowLogsPod] = useState<IKubePodSchema>()
    const [desiredShowKubeEventsPod, setDesiredShowKubeEventsPod] = useState<IKubePodSchema>()
    const [desiredShowMonitorPod, setDesiredShowMonitorPod] = useState<IKubePodSchema>()
    const [desiredShowTerminalPod, setDesiredShowTerminalPod] = useState<IKubePodSchema>()
    const [advancedLog, setAdvancedLog] = useState(false)
    let clusterName = cluster?.name
    if (clusterName_) {
        clusterName = clusterName_
    }
    const { yataiComponentsInfo } = useFetchYataiComponents(clusterName)

    const hasLogging = yataiComponentsInfo.data?.find((x) => x.type === 'logging') !== undefined
    const hasMonitoring = yataiComponentsInfo.data?.find((x) => x.type === 'monitoring') !== undefined

    const apiServerPods = pods?.filter((pod) => !pod.runner_name) ?? []

    const runnerPodsGroup =
        pods?.reduce((acc, pod) => {
            const { runner_name: runnerName } = pod
            if (!runnerName) {
                return acc
            }
            const pods_ = acc[runnerName] ?? []
            return {
                ...acc,
                [runnerName]: [...pods_, pod],
            }
        }, {} as Record<string, IKubePodSchema[]>) ?? {}

    const runnerNames = Object.keys(runnerPodsGroup).sort((a, b) => {
        return runnerPodsGroup[a][0].name.localeCompare(runnerPodsGroup[b][0].name)
    })

    const renderPodRow = useCallback(
        (pod: IKubePodSchema) => {
            return (
                <StyledTableBodyRow key={pod.name}>
                    <StyledTableBodyCell>
                        <StatefulTooltip key={pod.name} content={pod.name} showArrow>
                            <div
                                style={{
                                    display: 'inline-block',
                                    whiteSpace: 'nowrap',
                                    overflow: 'hidden',
                                    textOverflow: 'ellipsis',
                                    maxWidth: 320,
                                }}
                            >
                                {pod.name}
                            </div>
                        </StatefulTooltip>
                    </StyledTableBodyCell>
                    <StyledTableBodyCell>
                        <PodStatus key={pod.name} pod={pod} pods={pods} />
                    </StyledTableBodyCell>
                    <StyledTableBodyCell>{t(pod.pod_status.status)}</StyledTableBodyCell>
                    <StyledTableBodyCell>
                        {pod.deployment_target ? t(pod.deployment_target.type) : '-'}
                    </StyledTableBodyCell>
                    <StyledTableBodyCell>{pod.node_name}</StyledTableBodyCell>
                    <StyledTableBodyCell>{formatDateTime(pod.status.start_time)}</StyledTableBodyCell>
                    <StyledTableBodyCell>
                        <div
                            key={pod.name}
                            style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: 8,
                            }}
                        >
                            <StatefulTooltip content={t('view log')} showArrow>
                                <Button size='mini' shape='circle' onClick={() => setDesiredShowLogsPod(pod)}>
                                    <IoMdList />
                                </Button>
                            </StatefulTooltip>
                            <StatefulTooltip content={t('events')} showArrow>
                                <Button size='mini' shape='circle' onClick={() => setDesiredShowKubeEventsPod(pod)}>
                                    <MdEventNote />
                                </Button>
                            </StatefulTooltip>
                            <StatefulTooltip content={t('terminal')} showArrow>
                                <Button size='mini' shape='circle' onClick={() => setDesiredShowTerminalPod(pod)}>
                                    <GoTerminal />
                                </Button>
                            </StatefulTooltip>
                            {hasMonitoring ? (
                                <StatefulTooltip content={t('monitor')} showArrow>
                                    <Button
                                        disabled={!hasMonitoring}
                                        size='mini'
                                        shape='circle'
                                        onClick={() => setDesiredShowMonitorPod(pod)}
                                    >
                                        <AiOutlineDashboard />
                                    </Button>
                                </StatefulTooltip>
                            ) : (
                                <StatefulTooltip
                                    content={t('please install yatai component first', [t('monitoring')])}
                                    showArrow
                                >
                                    <div
                                        style={{
                                            display: 'flex',
                                            alignItems: 'center',
                                            gap: 2,
                                        }}
                                    >
                                        <Button
                                            disabled
                                            size='mini'
                                            shape='circle'
                                            onClick={() => setDesiredShowMonitorPod(pod)}
                                        >
                                            <AiOutlineDashboard />
                                        </Button>
                                        <div
                                            style={{
                                                cursor: 'pointer',
                                            }}
                                        >
                                            <AiOutlineQuestionCircle size={10} />
                                        </div>
                                    </div>
                                </StatefulTooltip>
                            )}
                        </div>
                    </StyledTableBodyCell>
                </StyledTableBodyRow>
            )
        },
        [hasMonitoring, pods, t]
    )

    return (
        <>
            <StyledTable>
                <StyledTableHeadRow>
                    {groupByRunner && <StyledTableHeadCell>{t('group')}</StyledTableHeadCell>}
                    <StyledTableHeadCell>{t('name')}</StyledTableHeadCell>
                    <StyledTableHeadCell>{t('status')}</StyledTableHeadCell>
                    <StyledTableHeadCell>{t('status name')}</StyledTableHeadCell>
                    <StyledTableHeadCell>{t('type')}</StyledTableHeadCell>
                    <StyledTableHeadCell>{t('node')}</StyledTableHeadCell>
                    <StyledTableHeadCell>{t('start time')}</StyledTableHeadCell>
                    <StyledTableHeadCell>{t('operation')}</StyledTableHeadCell>
                </StyledTableHeadRow>
                {loading ? (
                    <tbody>
                        <tr>
                            <td
                                colSpan={groupByRunner ? 8 : 7}
                                style={{
                                    padding: 20,
                                }}
                            >
                                <Skeleton animation rows={3} />
                            </td>
                        </tr>
                    </tbody>
                ) : (
                    <>
                        {groupByRunner && (
                            <>
                                <StyledTableBodyRow>
                                    <StyledTableBodyCell rowSpan={apiServerPods.length + 1}>
                                        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                                            {React.createElement(resourceIconMapping.bento_api_server, { size: 12 })}
                                            <span>API Server</span>
                                        </div>
                                    </StyledTableBodyCell>
                                </StyledTableBodyRow>
                                {apiServerPods.map(renderPodRow)}
                            </>
                        )}
                        {groupByRunner &&
                            runnerNames.reduce((acc, runnerName) => {
                                const pods_ = runnerPodsGroup[runnerName]
                                return [
                                    ...acc,
                                    <StyledTableBodyRow key={runnerName}>
                                        <StyledTableBodyCell rowSpan={pods_.length + 1}>
                                            <div
                                                style={{
                                                    display: 'flex',
                                                    flexDirection: 'column',
                                                    gap: 6,
                                                }}
                                            >
                                                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                                                    {React.createElement(resourceIconMapping.bento_runner, {
                                                        size: 12,
                                                    })}
                                                    <span>Runner</span>
                                                </div>
                                                <span style={{ fontWeight: 'bolder' }}>{runnerName}</span>
                                            </div>
                                        </StyledTableBodyCell>
                                    </StyledTableBodyRow>,
                                    ...pods_.map(renderPodRow),
                                ]
                            }, [] as React.ReactNode[])}
                        {!groupByRunner && pods.map(renderPodRow)}
                    </>
                )}
            </StyledTable>
            <Modal
                overrides={{
                    Dialog: {
                        style: {
                            width: '80vw',
                            height: '80vh',
                        },
                    },
                }}
                isOpen={desiredShowLogsPod !== undefined}
                onClose={() => setDesiredShowLogsPod(undefined)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>
                    <div
                        style={{
                            display: 'flex',
                            alignItems: 'center',
                        }}
                    >
                        <div style={{ marginRight: 40 }}>{t('view log')}</div>
                        <Label
                            style={{
                                fontSize: 12,
                            }}
                        >
                            {t('advanced')}
                        </Label>
                        <Toggle disabled={!hasLogging} value={advancedLog} onChange={setAdvancedLog} />
                    </div>
                </ModalHeader>
                <ModalBody>
                    {organization &&
                        clusterName &&
                        desiredShowLogsPod &&
                        (advancedLog ? (
                            <div style={{ height: 'calc(80vh - 100px)' }}>
                                <LokiLog podName={desiredShowLogsPod.name} namespace={desiredShowLogsPod.namespace} />
                            </div>
                        ) : (
                            <Log
                                open={desiredShowLogsPod !== undefined}
                                clusterName={clusterName}
                                deploymentName={deployment?.name}
                                namespace={desiredShowLogsPod.namespace}
                                podName={desiredShowLogsPod.name}
                                width='auto'
                                height='calc(80vh - 200px)'
                            />
                        ))}
                </ModalBody>
            </Modal>
            <Modal
                overrides={{
                    Dialog: {
                        style: {
                            width: '80vw',
                            height: '80vh',
                        },
                    },
                }}
                isOpen={desiredShowKubeEventsPod !== undefined}
                onClose={() => setDesiredShowKubeEventsPod(undefined)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('events')}</ModalHeader>
                <ModalBody>
                    {organization && clusterName && desiredShowKubeEventsPod && (
                        <KubePodEvents
                            open={desiredShowKubeEventsPod !== undefined}
                            clusterName={clusterName}
                            deploymentName={deployment?.name}
                            namespace={desiredShowKubeEventsPod.namespace}
                            podName={desiredShowKubeEventsPod.name}
                            width='auto'
                            height='calc(80vh - 200px)'
                        />
                    )}
                </ModalBody>
            </Modal>
            <Modal
                overrides={{
                    Dialog: {
                        style: {
                            width: '80vw',
                        },
                    },
                }}
                isOpen={desiredShowMonitorPod !== undefined}
                onClose={() => setDesiredShowMonitorPod(undefined)}
                closeable
                animate
                size='default'
                autoFocus
            >
                <ModalHeader>{t('monitor')}</ModalHeader>
                <ModalBody>
                    {organization && clusterName && desiredShowMonitorPod && <PodMonitor pod={desiredShowMonitorPod} />}
                </ModalBody>
            </Modal>
            <Modal
                overrides={{
                    Dialog: {
                        style: {
                            width: '80vw',
                            height: '80vh',
                            display: 'flex',
                            flexDirection: 'column',
                        },
                    },
                }}
                isOpen={desiredShowTerminalPod !== undefined}
                onClose={() => setDesiredShowTerminalPod(undefined)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('terminal')}</ModalHeader>
                <ModalBody style={{ flex: '1 1 0' }}>
                    {organization && clusterName && desiredShowTerminalPod && (
                        <Terminal
                            open={desiredShowTerminalPod !== undefined}
                            clusterName={clusterName}
                            deploymentName={deployment?.name}
                            namespace={desiredShowTerminalPod.namespace}
                            podName={desiredShowTerminalPod.name}
                            containerName={desiredShowTerminalPod.raw_status?.containerStatuses?.[0].name ?? ''}
                        />
                    )}
                </ModalBody>
            </Modal>
        </>
    )
}
