import React from 'react'
import useTranslation from '@/hooks/useTranslation'
import { Select } from 'baseui/select'
import { DeploymentSnapshotCanaryRuleType } from '@/schemas/deployment_snapshot'

interface IDeploymentSnapshotCanaryRuleTypeSelectorProps {
    value?: DeploymentSnapshotCanaryRuleType
    onChange?: (value: DeploymentSnapshotCanaryRuleType) => void
    excludes?: DeploymentSnapshotCanaryRuleType[]
}

export default function DeploymentSnapshotCanaryRuleTypeSelector({
    value,
    onChange,
    excludes,
}: IDeploymentSnapshotCanaryRuleTypeSelectorProps) {
    const [t] = useTranslation()
    return (
        <Select
            onChange={(params) => {
                if (!params.option) {
                    return
                }
                onChange?.(params.option.id as DeploymentSnapshotCanaryRuleType)
            }}
            value={[{ id: value }]}
            options={[
                {
                    id: 'weight',
                    label: t('weight'),
                },
                {
                    id: 'header',
                    label: t('header'),
                },
                {
                    id: 'cookie',
                    label: t('cookie'),
                },
            ].filter((x) => {
                if (!excludes || excludes.length === 0) {
                    return true
                }
                return excludes.indexOf(x.id as DeploymentSnapshotCanaryRuleType) < 0
            })}
        />
    )
}
