import { useCluster } from '@/hooks/useCluster'
import { useDeployment } from '@/hooks/useDeployment'
import { useOrganization } from '@/hooks/useOrganization'
import useTranslation from '@/hooks/useTranslation'
import { IKubePodSchema } from '@/schemas/kube_pod'
import { formatTime } from '@/utils/datetime'
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
import { PodStatus } from './PodsStatus'
import Table from './Table'
import Terminal from './Terminal'
import DeploymentKubeEvents from './DeploymentKubeEvents'
import Toggle from './Toggle'
import Label from './Label'
import LokiLog from './LokiLog'
import PodMonitor from './PodMonitor'

export interface IPodListProps {
    loading?: boolean
    pods: IKubePodSchema[]
}

export default ({ loading = false, pods }: IPodListProps) => {
    const [t] = useTranslation()
    const { organization } = useOrganization()
    const { cluster } = useCluster()
    const { deployment } = useDeployment()
    const [desiredShowLogsPod, setDesiredShowLogsPod] = useState<IKubePodSchema>()
    const [desiredShowKubeEventsPod, setDesiredShowKubeEventsPod] = useState<IKubePodSchema>()
    const [desiredShowMonitorPod, setDesiredShowMonitorPod] = useState<IKubePodSchema>()
    const [desiredShowTerminalPod, setDesiredShowTerminalPod] = useState<IKubePodSchema>()
    const [advancedLog, setAdvancedLog] = useState(false)
    const { yataiComponentsInfo } = useFetchYataiComponents(organization?.name, cluster?.name)

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
                    pod.name,
                    <PodStatus key={pod.name} pod={pod} pods={pods} />,
                    t(pod.pod_status.status),
                    pod.deployment_snapshot ? t(pod.deployment_snapshot.type) : '-',
                    pod.node_name,
                    formatTime(pod.status.start_time),
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
                        <Toggle value={advancedLog} onChange={setAdvancedLog} />
                    </div>
                </ModalHeader>
                <ModalBody>
                    {organization &&
                        cluster &&
                        deployment &&
                        desiredShowLogsPod &&
                        (advancedLog ? (
                            <div style={{ height: 'calc(80vh - 100px)' }}>
                                <LokiLog podName={desiredShowLogsPod.name} />
                            </div>
                        ) : (
                            <Log
                                open={desiredShowLogsPod !== undefined}
                                orgName={organization.name}
                                clusterName={cluster.name}
                                deploymentName={deployment.name}
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
                    {organization && cluster && deployment && desiredShowKubeEventsPod && (
                        <DeploymentKubeEvents
                            open={desiredShowKubeEventsPod !== undefined}
                            orgName={organization.name}
                            clusterName={cluster.name}
                            deploymentName={deployment.name}
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
                    {organization && cluster && deployment && desiredShowMonitorPod && (
                        <PodMonitor pod={desiredShowMonitorPod} />
                    )}
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
                    {organization && cluster && deployment && desiredShowTerminalPod && (
                        <Terminal
                            open={desiredShowTerminalPod !== undefined}
                            orgName={organization.name}
                            clusterName={cluster.name}
                            deploymentName={deployment.name}
                            podName={desiredShowTerminalPod.name}
                            containerName={desiredShowTerminalPod.raw_status?.containerStatuses?.[0].name ?? ''}
                        />
                    )}
                </ModalBody>
            </Modal>
        </>
    )
}
