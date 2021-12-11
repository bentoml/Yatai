/* eslint-disable jsx-a11y/no-static-element-interactions */
import React, { useCallback, useState } from 'react'
import { Tag, KIND as TagKind, VARIANT as TagVariant } from 'baseui/tag'
import { ImageBuildStatus } from '@/schemas/bento'
import { StyledSpinnerNext } from 'baseui/spinner'
import useTranslation from '@/hooks/useTranslation'
import { IKubePodSchema } from '@/schemas/kube_pod'
import { useFetchClusterPods } from '@/hooks/useFetchClusterPods'
import { StatefulPopover } from 'baseui/popover'
import { Button } from 'baseui/button'
import { VscDebugRerun } from 'react-icons/vsc'
import PodList from './PodList'
import Card from './Card'

const imageBuildStatusColorMap: Record<ImageBuildStatus, keyof TagKind> = {
    pending: TagKind.primary,
    building: TagKind.accent,
    failed: TagKind.negative,
    success: TagKind.positive,
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

export interface IBentoImageBuildStatusProps {
    status: ImageBuildStatus
    podsSelector?: string
    onRerunClick?: () => Promise<void>
}

export default function ImageBuildStatusTag({ status, podsSelector, onRerunClick }: IBentoImageBuildStatusProps) {
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
            <Tag closeable={false} variant={TagVariant.light} kind={imageBuildStatusColorMap[status]}>
                <div
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 4,
                    }}
                >
                    {['pending', 'building'].indexOf(status) >= 0 && <StyledSpinnerNext $size={100} />}
                    {t(status)}
                </div>
            </Tag>
        </StatefulPopover>
    )
}
