import React from 'react'
import { YataiComponentType } from '@/schemas/yatai_component'
import { Select } from 'baseui/select'
import YataiComponentTypeRender from './YataiComponentTypeRender'

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
    return (
        <Select
            options={(
                [
                    {
                        id: 'deployment',
                        label: <YataiComponentTypeRender type='deployment' />,
                    },
                    {
                        id: 'logging',
                        label: <YataiComponentTypeRender type='logging' />,
                    },
                    {
                        id: 'monitoring',
                        label: <YataiComponentTypeRender type='monitoring' />,
                    },
                ] as { id: YataiComponentType; label: React.ReactNode }[]
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
