import React from 'react'
import useTranslation from '@/hooks/useTranslation'
import { Select } from 'baseui/select'
import { DeploymentSnapshotType } from '@/schemas/deployment_snapshot'

export interface IDeploymentSnapshotTypeSelectorProps {
    value?: DeploymentSnapshotType
    onChange?: (newValue: DeploymentSnapshotType) => void
}

export default function DeploymentSnapshotTypeSelector({ value, onChange }: IDeploymentSnapshotTypeSelectorProps) {
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
                ] as { id: DeploymentSnapshotType; label: string }[]
            }
            onChange={(params) => {
                if (!params.option) {
                    return
                }
                onChange?.(params.option.id as DeploymentSnapshotType)
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
