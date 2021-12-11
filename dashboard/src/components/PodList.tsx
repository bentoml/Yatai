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
import React, { useState } from 'react'
import { StatefulTooltip } from 'baseui/tooltip'
import { Button } from 'baseui/button'
import { AiOutlineDashboard, AiOutlineQuestionCircle } from 'react-icons/ai'
import { useFetchYataiComponents } from '@/hooks/useFetchYataiComponents'
import Log from './Log'
import { PodStatus } from './PodStatuses'
import Table from './Table'
import Terminal from './Terminal'
import KubePodEvents from './KubePodEvents'
import Toggle from './Toggle'
import Label from './Label'
import LokiLog from './LokiLog'
import PodMonitor from './PodMonitor'

export interface IPodListProps {
    loading?: boolean
    clusterName?: string
    pods: IKubePodSchema[]
}

export default ({ loading = false, clusterName: clusterName_, pods }: IPodListProps) => {
    const [t] = useTranslation()
    const { organization } = useOrganization()
    const { cluster } = useCluster()
    const { deployment } = useDeployment()
    const [desiredShowLogsPod, setDesiredShowLogsPod] = useState<IKubePodSchema>()
    const [desiredShowKubeEventsPod, setDesiredShowKubeEventsPod] = useState<IKubePodSchema>()
    const [desiredShowMonitorPod, setDesiredShowMonitorPod] = useState<IKubePodSchema>()
    const [desiredShowTerminalPod, setDesiredShowTerminalPod] = useState<IKubePodSchema>()
    const [advancedLog, setAdvancedLog] = useState(false)
    let clusterName = clusterName_
    if (cluster) {
        clusterName = cluster.name
    }
    const { yataiComponentsInfo } = useFetchYataiComponents(clusterName)

    const hasLogging = yataiComponentsInfo.data?.find((x) => x.type === 'logging') !== undefined
    const hasMonitoring = yataiComponentsInfo.data?.find((x) => x.type === 'monitoring') !== undefined

    return (
        <>
            <Table
                isLoading={loading}
                columns={[
                    t('name'),
                    t('status'),
                    t('status name'),
                    t('type'),
                    t('node'),
                    t('start time'),
                    t('operation'),
                ]}
                data={pods.map((pod) => [
                    <StatefulTooltip key={pod.name} content={pod.name} showArrow>
                        <div
                            style={{
                                whiteSpace: 'nowrap',
                                overflow: 'hidden',
                                textOverflow: 'ellipsis',
                                maxWidth: 320,
                            }}
                        >
                            {pod.name}
                        </div>
                    </StatefulTooltip>,
                    <PodStatus key={pod.name} pod={pod} pods={pods} />,
                    t(pod.pod_status.status),
                    pod.deployment_target ? t(pod.deployment_target.type) : '-',
                    pod.node_name,
                    formatDateTime(pod.status.start_time),
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
                    </div>,
                ])}
            />
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
