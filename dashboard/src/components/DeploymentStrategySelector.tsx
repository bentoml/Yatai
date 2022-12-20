import useTranslation from '@/hooks/useTranslation'
import { DeploymentStrategy } from '@/schemas/deployment_target'
import { Select, SelectProps } from 'baseui/select'

export interface IDeploymentStrategySelectorProps {
    value?: DeploymentStrategy
    onChange?: (value: DeploymentStrategy) => void
    overrides?: SelectProps['overrides']
    disabled?: boolean
}

export default function DeploymentStrategySelector({
    value = 'RollingUpdate',
    onChange,
    overrides,
    disabled,
}: IDeploymentStrategySelectorProps) {
    const [t] = useTranslation()

    return (
        <Select
            disabled={disabled}
            overrides={overrides}
            options={
                [
                    {
                        id: 'RollingUpdate',
                        label: t('RollingUpdate'),
                    },
                    {
                        id: 'Recreate',
                        label: t('Recreate'),
                    },
                    {
                        id: 'RampedSlowRollout',
                        label: t('RampedSlowRollout'),
                    },
                    {
                        id: 'BestEffortControlledRollout',
                        label: t('BestEffortControlledRollout'),
                    },
                ] as { id: DeploymentStrategy; label: string }[]
            }
            onChange={(params) => {
                if (!params.option) {
                    return
                }
                onChange?.(params.option.id as DeploymentStrategy)
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
