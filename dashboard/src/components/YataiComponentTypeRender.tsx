import { yataiComponentIconMapping } from '@/consts'
import { YataiComponentType } from '@/schemas/yatai_component'
import React from 'react'

export interface IYataiComponentTypeProps {
    type: YataiComponentType
}

export default function YataiComponentTypeRender({ type }: IYataiComponentTypeProps) {
    return (
        <div
            style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: 6,
                fontSize: 14,
            }}
        >
            {React.createElement(yataiComponentIconMapping[type], { size: 16 })}
            <span>{type}</span>
        </div>
    )
}
