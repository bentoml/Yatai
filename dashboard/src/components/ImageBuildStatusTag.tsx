/* eslint-disable jsx-a11y/no-static-element-interactions */
import React, { useState } from 'react'
import { Tag, KIND as TagKind, VARIANT as TagVariant } from 'baseui/tag'
import { ImageBuildStatus } from '@/schemas/bento'
import { StyledSpinnerNext } from 'baseui/spinner'
import useTranslation from '@/hooks/useTranslation'
import { IKubePodSchema } from '@/schemas/kube_pod'
import { useFetchClusterPods } from '@/hooks/useFetchClusterPods'
import { useStyletron } from 'baseui'
import { StatefulPopover } from 'baseui/popover'
import PodList from './PodList'

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
}

export default function ImageBuildStatusTag({ status, podsSelector }: IBentoImageBuildStatusProps) {
    const [t] = useTranslation()

    const [, theme] = useStyletron()

    return (
        <StatefulPopover
            placement='bottomLeft'
            showArrow
            returnFocus
            autoFocus
            content={() => {
                return (
                    podsSelector && (
                        <div
                            style={{
                                boxShadow: theme.lighting.shadow400,
                            }}
                            onClick={(e) => e.stopPropagation()}
                        >
                            <Pods selector={podsSelector} />
                        </div>
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
