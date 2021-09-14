import { useCluster } from '@/hooks/useCluster'
import { useDeployment } from '@/hooks/useDeployment'
import { useOrganization } from '@/hooks/useOrganization'
import useTranslation from '@/hooks/useTranslation'
import { IKubePodSchema } from '@/schemas/kube_pod'
import { formatTime } from '@/utils/datetime'
import { Modal, ModalBody, ModalHeader } from 'baseui/modal'
import { IoMdList } from 'react-icons/io'
import { GoTerminal } from 'react-icons/go'
import React, { useState } from 'react'
import { StatefulTooltip } from 'baseui/tooltip'
import { Button } from 'baseui/button'
import Log from './Log'
import { PodStatus } from './PodsStatus'
import Table from './Table'
import Terminal from './Terminal'

export interface IPodListProps {
    loading?: boolean
    pods: IKubePodSchema[]
}

export default ({ loading = false, pods }: IPodListProps) => {
    const [t] = useTranslation()
    const { organization } = useOrganization()
    const { cluster } = useCluster()
    const { deployment } = useDeployment()
    const [desiredShowLogPod, setDesiredShowLogPod] = useState<IKubePodSchema>()
    const [desiredShowTerminalPod, setDesiredShowTerminalPod] = useState<IKubePodSchema>()

    return (
        <>
            <Table
                isLoading={loading}
                columns={[t('name'), t('status'), t('status name'), t('start time'), t('operation')]}
                data={pods.map((pod) => [
                    pod.name,
                    <PodStatus key={pod.name} pod={pod} pods={pods} />,
                    t(pod.pod_status.status),
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
                            <Button size='mini' shape='circle' onClick={() => setDesiredShowLogPod(pod)}>
                                <IoMdList />
                            </Button>
                        </StatefulTooltip>
                        <StatefulTooltip content={t('terminal')} showArrow>
                            <Button size='mini' shape='circle' onClick={() => setDesiredShowTerminalPod(pod)}>
                                <GoTerminal />
                            </Button>
                        </StatefulTooltip>
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
                isOpen={desiredShowLogPod !== undefined}
                onClose={() => setDesiredShowLogPod(undefined)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('view log')}</ModalHeader>
                <ModalBody>
                    {organization && cluster && deployment && desiredShowLogPod && (
                        <Log
                            open={desiredShowLogPod !== undefined}
                            orgName={organization.name}
                            clusterName={cluster.name}
                            deploymentName={deployment.name}
                            podName={desiredShowLogPod.name}
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
