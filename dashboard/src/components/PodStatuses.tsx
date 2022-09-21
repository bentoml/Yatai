import React, { useState } from 'react'
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
    onClick?: () => void
}

export const PodStatus = React.memo(
    ({ pod, onClick, pods }: IPodStatusProps) => {
        const idx = pods.findIndex((x) => x.name === pod.name)
        const styles = useStyles()
        const isCreating = pod.pod_status.status === 'Pending'
        const isTerminating = pod.pod_status.status === 'Terminating'
        const isYellow = pod.status.is_old && ['Running', 'Pending', 'Succeeded'].indexOf(pod.pod_status.status) >= 0
        const isGreen = !isYellow && ['Running', 'Succeeded'].indexOf(pod.pod_status.status) >= 0
        const [showLogModal, setShowLogModal] = useState(false)
        const { organization } = useOrganization()
        const { cluster } = useCluster()
        const { deployment } = useDeployment()

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
                        {pod.pod_status.status === 'Succeeded' ? <AiFillCheckCircle /> : <div />}
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
                    isOpen={showLogModal}
                    onClose={() => setShowLogModal(false)}
                    closeable
                    animate
                    autoFocus
                >
                    <ModalHeader>{t('view log')}</ModalHeader>
                    <ModalBody>
                        {organization && cluster && deployment && (
                            <Log
                                open={showLogModal}
                                clusterName={cluster.name}
                                namespace={pod.namespace}
                                deploymentName={deployment.name}
                                podName={pod.name}
                                width='auto'
                                height='calc(80vh - 200px)'
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

interface IPodsStatusProps {
    pods: IKubePodSchema[]
    replicas: number
    // eslint-disable-next-line react/require-default-props
    style?: React.CSSProperties
}

const PodStatuses = ({ pods, replicas, style }: IPodsStatusProps) => {
    const styles = useStyles()
    const lacking = replicas - pods.length
    return (
        <div className={styles.ballContainer} style={style}>
            {pods.map((pod, idx) => {
                return <PodStatus key={idx} pod={pod} pods={pods} />
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
    ({ pods, replicas, style }: IPodsStatusProps) => {
        // TODO: split once
        const productionPods = pods.filter((pod) => !pod.status.is_canary)
        const canaryPods = pods.filter((pod) => pod.status.is_canary)
        return (
            <div style={style}>
                <PodStatuses pods={productionPods} replicas={replicas} />
                {canaryPods.length > 0 && (
                    <PodStatuses pods={canaryPods} replicas={replicas} style={{ marginTop: 10 }} />
                )}
            </div>
        )
    },
    (prevProps, nextProps) => {
        return _.isEqual(prevProps, nextProps)
    }
)
