import { IResourceSchema } from '@/schemas/resource'
import { Tag } from 'baseui/tag'
import React from 'react'

export interface IResourceLabelsProps {
    resource: IResourceSchema
}

export function ResourceLabels({ resource }: IResourceLabelsProps) {
    return (
        <div
            style={{
                display: 'flex',
                alignItems: 'center',
                flexWrap: 'wrap',
                gap: 4,
            }}
        >
            {resource.labels.map((label) => {
                const labelText = label.value === '' ? label.key : `${label.key}: ${label.value}`
                return (
                    <Tag
                        overrides={{
                            Root: {
                                style: {
                                    margin: '0px !important',
                                    fontSize: '12px !important',
                                },
                            },
                        }}
                        key={label.key}
                        kind='purple'
                        closeable={false}
                        variant='solid'
                        size='small'
                    >
                        {labelText}
                    </Tag>
                )
            })}
        </div>
    )
}
