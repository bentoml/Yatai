import React from 'react'
import useTranslation from '@/hooks/useTranslation'
import { Select } from 'baseui/select'
import { DeploymentTargetType } from '@/schemas/deployment_target'

export interface IDeploymentTargetTypeSelectorProps {
    value?: DeploymentTargetType
    onChange?: (newValue: DeploymentTargetType) => void
}

export default function DeploymentTargetTypeSelector({ value, onChange }: IDeploymentTargetTypeSelectorProps) {
    const [t] = useTranslation()

    return (
        <Select
            options={
                [
                    {
                        id: 'stable',
                        label: t('stable'),
                    },
                    {
                        id: 'canary',
                        label: t('canary'),
                    },
                ] as { id: DeploymentTargetType; label: string }[]
            }
            onChange={(params) => {
                if (!params.option) {
                    return
                }
                onChange?.(params.option.id as DeploymentTargetType)
            }}
            value={
                value
                    ? [
                          {
                              id: value,
                          },
                      ]
                    : []
            }
        />
    )
}
