import React from 'react'
import { Tag, KIND as TagKind, VARIANT as TagVariant } from 'baseui/tag'
import { ModelVersionImageBuildStatus } from '@/schemas/model_version'
import { StyledSpinnerNext } from 'baseui/spinner'
import useTranslation from '@/hooks/useTranslation'

const imageBuildStatusColorMap: Record<ModelVersionImageBuildStatus, keyof TagKind> = {
    pending: TagKind.primary,
    building: TagKind.accent,
    failed: TagKind.negative,
    success: TagKind.positive,
}

export interface IModelVersionImageBuildStatusProps {
    status: ModelVersionImageBuildStatus
}

export default function ModelVersionImageBuildStatusTag({ status }: IModelVersionImageBuildStatusProps) {
    const [t] = useTranslation()
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
                {t(status)}
            </div>
        </Tag>
    )
}
