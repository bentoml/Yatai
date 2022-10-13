import { useCluster } from '@/hooks/useCluster'
import { useOrganization } from '@/hooks/useOrganization'
import useTranslation from '@/hooks/useTranslation'
import { IKubePodSchema } from '@/schemas/kube_pod'
import { Modal, ModalBody, ModalHeader } from 'baseui/modal'
import { IoMdList } from 'react-icons/io'
import { GoTerminal } from 'react-icons/go'
import { MdEventNote } from 'react-icons/md'
import React, { useState, useCallback, useMemo } from 'react'
import { StatefulTooltip } from 'baseui/tooltip'
import { Button } from 'baseui/button'
import {
    StyledTable,
    StyledTableBodyCell,
    StyledTableBodyRow,
    StyledTableHeadCell,
    StyledTableHeadRow,
} from 'baseui/table-semantic'
import { Skeleton } from 'baseui/skeleton'
import { resourceIconMapping } from '@/consts'
import { IDeploymentSchema } from '@/schemas/deployment'
import Log from './Log'
import { PodStatus } from './PodStatuses'
import Terminal from './Terminal'
import KubePodEvents from './KubePodEvents'
import PodMonitor from './PodMonitor'
import Time from './Time'

export interface IPodListProps {
    loading?: boolean
    clusterName?: string
    groupByRunner?: boolean
    pods: IKubePodSchema[]
    deployment?: IDeploymentSchema
}

export default function PodList({
    loading = false,
    deployment,
    clusterName: clusterName_,
    pods,
    groupByRunner = false,
}: IPodListProps) {
    const [t] = useTranslation()
    const { organization } = useOrganization()
    const { cluster } = useCluster()
    const [desiredShowLogsPodName, setDesiredShowLogsPodName] = useState<string>()
    const desiredShowLogsPod = useMemo(
        () => pods.find((pod) => pod.name === desiredShowLogsPodName),
        [desiredShowLogsPodName, pods]
    )
    const [desiredShowKubeEventsPod, setDesiredShowKubeEventsPod] = useState<IKubePodSchema>()
    const [desiredShowMonitorPod, setDesiredShowMonitorPod] = useState<IKubePodSchema>()
    const [desiredShowTerminalPod, setDesiredShowTerminalPod] = useState<IKubePodSchema>()
    let clusterName = cluster?.name
    if (clusterName_) {
        clusterName = clusterName_
    }

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
                    <StyledTableBodyCell>
                        <Time time={pod.status.start_time} />
                    </StyledTableBodyCell>
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
                                <Button size='mini' shape='circle' onClick={() => setDesiredShowLogsPodName(pod.name)}>
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
                        </div>
                    </StyledTableBodyCell>
                </StyledTableBodyRow>
            )
        },
        [pods, t]
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
                isOpen={desiredShowLogsPodName !== undefined}
                onClose={() => setDesiredShowLogsPodName(undefined)}
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
                    </div>
                </ModalHeader>
                <ModalBody>
                    {organization && clusterName && desiredShowLogsPod ? (
                        <Log
                            open={desiredShowLogsPod !== undefined}
                            clusterName={clusterName}
                            deploymentName={deployment?.name}
                            pod={desiredShowLogsPod}
                            width='auto'
                            height='calc(80vh - 200px)'
                        />
                    ) : (
                        t('pod {{0}} is not exists', [desiredShowLogsPodName])
                    )}
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
                            containerName={
                                desiredShowTerminalPod.raw_status?.containerStatuses?.[
                                    (desiredShowTerminalPod.raw_status?.containerStatuses?.length ?? 1) - 1
                                ]?.name ?? ''
                            }
                        />
                    )}
                </ModalBody>
            </Modal>
        </>
    )
}
