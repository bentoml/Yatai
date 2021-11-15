import React from 'react'
import useTranslation from '@/hooks/useTranslation'
import { Select } from 'baseui/select'
import { DeploymentTargetCanaryRuleType } from '@/schemas/deployment_target'

interface IDeploymentTargetCanaryRuleTypeSelectorProps {
    value?: DeploymentTargetCanaryRuleType
    onChange?: (value: DeploymentTargetCanaryRuleType) => void
    excludes?: DeploymentTargetCanaryRuleType[]
}

export default function DeploymentTargetCanaryRuleTypeSelector({
    value,
    onChange,
    excludes,
}: IDeploymentTargetCanaryRuleTypeSelectorProps) {
    const [t] = useTranslation()
    return (
        <Select
            onChange={(params) => {
                if (!params.option) {
                    return
                }
                onChange?.(params.option.id as DeploymentTargetCanaryRuleType)
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
                return excludes.indexOf(x.id as DeploymentTargetCanaryRuleType) < 0
            })}
        />
    )
}
