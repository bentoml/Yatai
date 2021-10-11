import { yataiComponentIconMapping } from '@/consts'
import useTranslation from '@/hooks/useTranslation'
import { YataiComponentType } from '@/schemas/yatai_component'
import React from 'react'

export interface IYataiComponentTypeProps {
    type: YataiComponentType
}

export default function YataiComponentTypeRender({ type }: IYataiComponentTypeProps) {
    const [t] = useTranslation()
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
            <span>{t(type)}</span>
        </div>
    )
}
