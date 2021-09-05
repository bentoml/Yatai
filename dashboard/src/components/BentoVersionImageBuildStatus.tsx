import React from 'react'
import { Tag, KIND as TagKind, VARIANT as TagVariant } from 'baseui/tag'
import { BentoVersionImageBuildStatus } from '@/schemas/bento_version'
import { StyledSpinnerNext } from 'baseui/spinner'

const imageBuildStatusColorMap: Record<BentoVersionImageBuildStatus, keyof TagKind> = {
    pending: TagKind.primary,
    building: TagKind.accent,
    failed: TagKind.negative,
    success: TagKind.positive,
}

export interface IBentoVersionImageBuildStatusProps {
    status: BentoVersionImageBuildStatus
}

export default function BentoVersionImageBuildStatusTag({ status }: IBentoVersionImageBuildStatusProps) {
    return (
        <Tag closeable={false} variant={TagVariant.light} kind={imageBuildStatusColorMap[status]}>
            <div
                style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 4,
                }}
            >
                {['pending', 'building'].indexOf(status) >= 0 && <StyledSpinnerNext $size={100} />}
                {status}
            </div>
        </Tag>
    )
}
