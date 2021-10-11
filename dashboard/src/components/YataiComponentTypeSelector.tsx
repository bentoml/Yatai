import React from 'react'
import useTranslation from '@/hooks/useTranslation'
import { YataiComponentType } from '@/schemas/yatai_component'
import { Select } from 'baseui/select'

export interface IYataiComponentTypeSelectorProps {
    value?: YataiComponentType
    onChange?: (newValue: YataiComponentType) => void
    excludes?: YataiComponentType[]
}

export default function YataiComponentTypeSelector({
    value,
    onChange,
    excludes = [],
}: IYataiComponentTypeSelectorProps) {
    const [t] = useTranslation()

    return (
        <Select
            options={(
                [
                    {
                        id: 'logging',
                        label: t('logging'),
                    },
                    {
                        id: 'monitoring',
                        label: t('monitoring'),
                    },
                ] as { id: YataiComponentType; label: string }[]
            ).filter((x) => excludes.indexOf(x.id) < 0)}
            onChange={(params) => {
                if (!params.option) {
                    return
                }
                onChange?.(params.option.id as YataiComponentType)
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
