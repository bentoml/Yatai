import React, { useMemo, useState } from 'react'
import { createUseStyles } from 'react-jss'
import { StatefulTooltip } from 'baseui/tooltip'
import classNames from 'classnames'
import _ from 'lodash'
import { IKubePodSchema } from '@/schemas/kube_pod'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import useTranslation from '@/hooks/useTranslation'
import { useOrganization } from '@/hooks/useOrganization'
import { useDeployment } from '@/hooks/useDeployment'
import { useCluster } from '@/hooks/useCluster'
import { AiFillCheckCircle } from 'react-icons/ai'
import { IoMdList } from 'react-icons/io'
import { Button } from 'baseui/button'
import { MdEventNote } from 'react-icons/md'
import { StatefulPopover } from 'baseui/popover'
import { StatefulMenu } from 'baseui/menu'
import { TbBrandDocker } from 'react-icons/tb'
import { GoTerminal } from 'react-icons/go'
import { VscDebugConsole } from 'react-icons/vsc'
import Terminal from './Terminal'
import PodMonitor from './PodMonitor'
import KubePodEvents from './KubePodEvents'
import Log from './Log'

export const useStyles = createUseStyles({
    '@keyframes creating': {
        '50%': {
            background: '#deecf9',
        },
        '100%': {
            background: '#269afd',
        },
    },
    '@keyframes terminating': {
        '50%': {
            background: '#ffffff',
        },
        '100%': {
            background: '#e3008c',
        },
    },
    'ballContainer': {
        display: 'flex',
    },
    'ball': {
        'cursor': 'pointer',
        '& > *': {
            display: 'flex',
            marginRight: 4,
        },
        '& > div': {
            borderRadius: '50%',
            width: 12,
            height: 12,
        },
    },
    'canaryGreen': {
        '& > div': {
            background: '#9be2be',
        },
        '& > svg': {
            color: '#9be2be',
        },
    },
    'green': {
        '& > div': {
            background: '#00ad56',
        },
        '& > svg': {
            color: '#00ad56',
        },
    },
    'canaryRed': {
        '& > div': {
            background: '#ec7e81',
        },
        '& > svg': {
            color: '#ec7e81',
        },
    },
    'red': {
        '& > div': {
            background: '#d13438',
        },
        '& > svg': {
            color: '#d13438',
        },
    },
    'yellow': {
        '& > div': {
            background: '#ffaa44',
        },
        '& > svg': {
            color: '#ffaa44',
        },
    },
    'grey': {
        '& > div': {
            background: '#a0aeb2',
        },
        '& > svg': {
            color: '#a0aeb2',
        },
    },
    'black': {
        '& > div': {
            background: '#000',
        },
        '& > svg': {
            color: '#000',
        },
    },
    'phantomGreen': {
        '& > div': {
            background:
                'linear-gradient(45deg, #ace8ca 25%, #00ad56 25%, #00ad56 50%, #ace8ca 50%, #ace8ca 75%, #00ad56 75%, #00ad56 100%)',
        },
        '& > svg': {
            color: 'linear-gradient(45deg, #ace8ca 25%, #00ad56 25%, #00ad56 50%, #ace8ca 50%, #ace8ca 75%, #00ad56 75%, #00ad56 100%)',
        },
    },
    'phantomRed': {
        '& > div': {
            background:
                'linear-gradient(45deg, #e08f91 25%, #d13438 25%, #d13438 50%, #e08f91 50%, #e08f91 75%, #d13438 75%, #d13438 100%)',
        },
        '& > svg': {
            color: 'linear-gradient(45deg, #e08f91 25%, #d13438 25%, #d13438 50%, #e08f91 50%, #e08f91 75%, #d13438 75%, #d13438 100%)',
        },
    },
    'creating': {
        '& > div': {
            animationName: '$creating',
            animationIterationCount: 'infinite',
            animationDuration: '1s',
        },
        '& > svg': {
            animationName: '$creating',
            animationIterationCount: 'infinite',
            animationDuration: '1s',
        },
    },
    'terminating': {
        '& > div': {
            animationName: '$terminating',
            animationIterationCount: 'infinite',
            animationDuration: '1s',
        },
        '& > svg': {
            animationName: '$terminating',
            animationIterationCount: 'infinite',
            animationDuration: '1s',
        },
    },
})

interface IPodStatusProps {
    pod: IKubePodSchema
    pods: IKubePodSchema[]
    showOperationIcons?: boolean
    onClick?: () => void
}

export const PodStatus = React.memo(
    ({ pod, onClick, pods, showOperationIcons = false }: IPodStatusProps) => {
        const idx = pods.findIndex((x) => x.name === pod.name)
        const styles = useStyles()
        const isCreating = pod.pod_status.status === 'Pending'
        const isTerminating = pod.pod_status.status === 'Terminating'
        const isYellow = pod.status.is_old && ['Running', 'Pending', 'Succeeded'].indexOf(pod.pod_status.status) >= 0
        const isGreen = !isYellow && ['Running', 'Succeeded'].indexOf(pod.pod_status.status) >= 0
        const { organization } = useOrganization()
        const { cluster } = useCluster()
        const { deployment } = useDeployment()

        const [desiredShowLogsPodName, setDesiredShowLogsPodName] = useState<string>()
        const desiredShowLogsPod = useMemo(
            () => pods.find((pod_) => pod_.name === desiredShowLogsPodName),
            [desiredShowLogsPodName, pods]
        )
        const [desiredShowKubeEventsPod, setDesiredShowKubeEventsPod] = useState<IKubePodSchema>()
        const [desiredShowMonitorPod, setDesiredShowMonitorPod] = useState<IKubePodSchema>()
        const [desiredShowTerminalPod, setDesiredShowTerminalPod] = useState<IKubePodSchema>()
        const [desiredShowTerminalContainerName, setDesiredShowTerminalContainerName] = useState<string>()
        const [desiredShowDebugTerminalPod, setDesiredShowDebugTerminalPod] = useState<IKubePodSchema>()
        const [desiredShowDebugTerminalContainerName, setDesiredShowDebugTerminalContainerName] = useState<string>()
        const clusterName = cluster?.name

        const isRed = !isYellow && pod.pod_status.status === 'Failed'

        const msg = (pod.warnings ?? []).map((e) => e.message).join('\n')
        let greenClassName = styles.green

        if (pod.status.is_canary) {
            greenClassName = styles.canaryGreen
        }

        let redClassName = styles.red

        if (pod.status.is_canary) {
            redClassName = styles.canaryRed
        }

        const [t] = useTranslation()

        const el = (
            <>
                <StatefulTooltip
                    content={
                        <div>
                            <div style={{ marginBottom: 8 }}>
                                <span>{pod.name}</span>
                            </div>
                            {msg}
                            {showOperationIcons && (
                                <div
                                    key={pod.name}
                                    style={{
                                        display: 'flex',
                                        alignItems: 'center',
                                        gap: 8,
                                        marginTop: 10,
                                    }}
                                >
                                    <StatefulTooltip content={t('view log')} showArrow>
                                        <Button
                                            size='mini'
                                            shape='circle'
                                            onClick={() => setDesiredShowLogsPodName(pod.name)}
                                        >
                                            <IoMdList />
                                        </Button>
                                    </StatefulTooltip>
                                    <StatefulTooltip content={t('events')} showArrow>
                                        <Button
                                            size='mini'
                                            shape='circle'
                                            onClick={() => setDesiredShowKubeEventsPod(pod)}
                                        >
                                            <MdEventNote />
                                        </Button>
                                    </StatefulTooltip>
                                    <StatefulTooltip content={t('terminal')} showArrow>
                                        <StatefulPopover
                                            focusLock
                                            placement='bottom'
                                            overrides={{
                                                Inner: {
                                                    style: {
                                                        minWith: '200px',
                                                    },
                                                },
                                            }}
                                            content={({ close }) => (
                                                <StatefulMenu
                                                    items={
                                                        pod.raw_status?.containerStatuses?.map((container) => ({
                                                            label: (
                                                                <div
                                                                    style={{
                                                                        display: 'flex',
                                                                        alignItems: 'center',
                                                                        gap: 8,
                                                                    }}
                                                                >
                                                                    <div style={{ flexShrink: 0 }}>
                                                                        <TbBrandDocker size={14} />
                                                                    </div>
                                                                    {container.name}
                                                                </div>
                                                            ),
                                                            containerName: container.name,
                                                        })) ?? []
                                                    }
                                                    onItemSelect={({ item }) => {
                                                        setDesiredShowTerminalContainerName(item?.containerName)
                                                        setDesiredShowTerminalPod(pod)
                                                        close()
                                                    }}
                                                    overrides={{
                                                        List: { style: { height: '150px', width: '138px' } },
                                                    }}
                                                />
                                            )}
                                        >
                                            <Button size='mini' shape='circle'>
                                                <GoTerminal />
                                            </Button>
                                        </StatefulPopover>
                                    </StatefulTooltip>
                                    <StatefulTooltip content={t('debug')} showArrow>
                                        <StatefulPopover
                                            focusLock
                                            placement='bottom'
                                            overrides={{
                                                Inner: {
                                                    style: {
                                                        minWith: '200px',
                                                    },
                                                },
                                            }}
                                            content={({ close }) => (
                                                <StatefulMenu
                                                    items={
                                                        pod.raw_status?.containerStatuses?.map((container) => ({
                                                            label: (
                                                                <div
                                                                    style={{
                                                                        display: 'flex',
                                                                        alignItems: 'center',
                                                                        gap: 8,
                                                                    }}
                                                                >
                                                                    <div style={{ flexShrink: 0 }}>
                                                                        <TbBrandDocker size={14} />
                                                                    </div>
                                                                    {container.name}
                                                                </div>
                                                            ),
                                                            containerName: container.name,
                                                        })) ?? []
                                                    }
                                                    onItemSelect={({ item }) => {
                                                        setDesiredShowDebugTerminalContainerName(item?.containerName)
                                                        setDesiredShowDebugTerminalPod(pod)
                                                        close()
                                                    }}
                                                    overrides={{
                                                        List: { style: { height: '150px', width: '138px' } },
                                                    }}
                                                />
                                            )}
                                        >
                                            <Button size='mini' shape='circle'>
                                                <VscDebugConsole />
                                            </Button>
                                        </StatefulPopover>
                                    </StatefulTooltip>
                                </div>
                            )}
                        </div>
                    }
                    showArrow
                >
                    <div
                        role='button'
                        tabIndex={0}
                        onClick={onClick}
                        key={idx}
                        style={{
                            display: 'inline-block',
                        }}
                        className={classNames({
                            [styles.ball]: true,
                            [styles.yellow]: isYellow,
                            [greenClassName]: isGreen,
                            [styles.creating]: isCreating,
                            [styles.terminating]: isTerminating,
                            [redClassName]: isRed,
                        })}
                    >
                        {pod.pod_status.status === 'Succeeded' ? <AiFillCheckCircle size='13px' /> : <div />}
                    </div>
                </StatefulTooltip>
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
                        {organization && clusterName && desiredShowMonitorPod && (
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
                    isOpen={desiredShowTerminalPod !== undefined && desiredShowTerminalContainerName !== undefined}
                    onClose={() => {
                        setDesiredShowTerminalContainerName(undefined)
                        setDesiredShowTerminalPod(undefined)
                    }}
                    closeable
                    animate
                    autoFocus
                >
                    <ModalHeader>{`${t('terminal')} - ${
                        desiredShowTerminalPod?.name
                    } - ${desiredShowTerminalContainerName}`}</ModalHeader>
                    <ModalBody style={{ flex: '1 1 0' }}>
                        {organization && clusterName && desiredShowTerminalPod && desiredShowTerminalContainerName && (
                            <Terminal
                                open={
                                    desiredShowTerminalPod !== undefined &&
                                    desiredShowTerminalContainerName !== undefined
                                }
                                clusterName={clusterName}
                                deploymentName={deployment?.name}
                                namespace={desiredShowTerminalPod.namespace}
                                podName={desiredShowTerminalPod.name}
                                containerName={desiredShowTerminalContainerName ?? ''}
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
                    isOpen={
                        desiredShowDebugTerminalPod !== undefined && desiredShowDebugTerminalContainerName !== undefined
                    }
                    onClose={() => {
                        setDesiredShowDebugTerminalContainerName(undefined)
                        setDesiredShowDebugTerminalPod(undefined)
                    }}
                    closeable
                    animate
                    autoFocus
                >
                    <ModalHeader>{`${t('debug')} - ${
                        desiredShowDebugTerminalPod?.name
                    } - ${desiredShowDebugTerminalContainerName}`}</ModalHeader>
                    <ModalBody style={{ flex: '1 1 0' }}>
                        {organization &&
                            clusterName &&
                            desiredShowDebugTerminalPod &&
                            desiredShowDebugTerminalContainerName && (
                                <Terminal
                                    open={
                                        desiredShowDebugTerminalPod !== undefined &&
                                        desiredShowDebugTerminalContainerName !== undefined
                                    }
                                    clusterName={clusterName}
                                    deploymentName={deployment?.name}
                                    namespace={desiredShowDebugTerminalPod.namespace}
                                    podName={desiredShowDebugTerminalPod.name}
                                    containerName={desiredShowDebugTerminalContainerName ?? ''}
                                    debug
                                />
                            )}
                    </ModalBody>
                </Modal>
            </>
        )

        return el
    },
    (prevProps, nextProps) => {
        return _.isEqual(prevProps, nextProps)
    }
)

interface IPodStatusesProps {
    pods: IKubePodSchema[]
    replicas: number
    // eslint-disable-next-line react/require-default-props
    style?: React.CSSProperties
    // eslint-disable-next-line react/require-default-props
    showOperationIcons?: boolean
}

const PodStatuses = ({ pods, replicas, style, showOperationIcons = false }: IPodStatusesProps) => {
    const styles = useStyles()
    const lacking = replicas - pods.length
    return (
        <div className={styles.ballContainer} style={style}>
            {pods.map((pod, idx) => {
                return <PodStatus key={idx} pod={pod} pods={pods} showOperationIcons={showOperationIcons} />
            })}
            {lacking > 0 &&
                !pods[0]?.status.is_canary &&
                new Array(lacking).fill(0).map((__, idx) => {
                    return (
                        <div
                            key={pods.length + idx}
                            className={classNames({
                                [styles.ball]: true,
                                [styles.grey]: true,
                            })}
                        />
                    )
                })}
        </div>
    )
}

export default React.memo(
    ({ pods, replicas, style, showOperationIcons = false }: IPodStatusesProps) => {
        // TODO: split once
        const productionPods = pods.filter((pod) => !pod.status.is_canary)
        const canaryPods = pods.filter((pod) => pod.status.is_canary)
        return (
            <div style={style}>
                <PodStatuses pods={productionPods} replicas={replicas} showOperationIcons={showOperationIcons} />
                {canaryPods.length > 0 && (
                    <PodStatuses
                        pods={canaryPods}
                        replicas={replicas}
                        style={{ marginTop: 10 }}
                        showOperationIcons={showOperationIcons}
                    />
                )}
            </div>
        )
    },
    (prevProps, nextProps) => {
        return _.isEqual(prevProps, nextProps)
    }
)
