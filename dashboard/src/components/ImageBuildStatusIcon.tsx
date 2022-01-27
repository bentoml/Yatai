/* eslint-disable jsx-a11y/no-static-element-interactions */
import React, { useCallback, useState } from 'react'
import { ImageBuildStatus } from '@/schemas/bento'
import useTranslation from '@/hooks/useTranslation'
import { IKubePodSchema } from '@/schemas/kube_pod'
import { useFetchClusterPods } from '@/hooks/useFetchClusterPods'
import { StatefulPopover } from 'baseui/popover'
import { Button } from 'baseui/button'
import { VscDebugRerun } from 'react-icons/vsc'
import { IconBaseProps } from 'react-icons'
import { FcHighPriority, FcOk, FcOvertime, FcWorkflow } from 'react-icons/fc'
import { createUseStyles } from 'react-jss'
import PodList from './PodList'
import Card from './Card'

const useStyles = createUseStyles({
    '@keyframes spin': {
        '100%': {
            transform: 'rotate(360deg)',
        },
    },
    'spinner': {
        animation: '$spin 3s linear infinite',
    },
})

const imageBuildStatusIconMapping: Record<ImageBuildStatus, React.ComponentType<IconBaseProps>> = {
    pending: FcOvertime,
    building: FcWorkflow,
    failed: FcHighPriority,
    success: FcOk,
}

interface IPodsProps {
    selector: string
}

function Pods({ selector }: IPodsProps) {
    const [pods, setPods] = useState<IKubePodSchema[]>([])
    const [podsLoading, setPodsLoading] = useState(false)
    useFetchClusterPods({
        clusterName: 'default',
        namespace: 'yatai-builders',
        selector,
        setPods,
        setPodsLoading,
    })

    return <PodList clusterName='default' loading={podsLoading} pods={pods} />
}

export interface IImageBuildStatusIconProps {
    status: ImageBuildStatus
    podsSelector?: string
    onRerunClick?: () => Promise<void>
    size?: number
}

export default function ImageBuildStatusIcon({
    status,
    podsSelector,
    onRerunClick,
    size = 28,
}: IImageBuildStatusIconProps) {
    const styles = useStyles()
    const [t] = useTranslation()
    const [rerunLoading, setRerunLoading] = useState(false)

    const handleRerunClick = useCallback(async () => {
        setRerunLoading(true)
        try {
            await onRerunClick?.()
        } finally {
            setRerunLoading(false)
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    return (
        <div
            style={{
                display: 'inline-flex',
            }}
            onClick={(e) => {
                e.stopPropagation()
                e.preventDefault()
            }}
        >
            <StatefulPopover
                placement='bottomLeft'
                showArrow
                returnFocus
                autoFocus
                content={() => {
                    return (
                        podsSelector && (
                            <Card
                                style={{
                                    margin: 0,
                                }}
                                extra={
                                    onRerunClick && (
                                        <Button
                                            startEnhancer={<VscDebugRerun />}
                                            size='compact'
                                            onClick={handleRerunClick}
                                            isLoading={rerunLoading}
                                        >
                                            {t('rerun')}
                                        </Button>
                                    )
                                }
                            >
                                <Pods selector={podsSelector} />
                            </Card>
                        )
                    )
                }}
            >
                <div
                    style={{
                        display: 'inline-flex',
                        cursor: 'pointer',
                    }}
                >
                    {React.createElement(imageBuildStatusIconMapping[status], {
                        size,
                        className: status === 'building' ? styles.spinner : '',
                    })}
                </div>
            </StatefulPopover>
        </div>
    )
}
