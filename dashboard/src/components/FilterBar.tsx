import React from 'react'
import { useQ } from '@/hooks/useQ'
import useTranslation from '@/hooks/useTranslation'
import { useStyletron } from 'baseui'
import { Button } from 'baseui/button'
import { AiOutlineClear } from 'react-icons/ai'
import Filter, { IFilterProps } from './Filter'

export interface IFilterBarProps {
    prefix?: React.ReactNode
    filters: IFilterProps[]
}

export default function FilterBar({ prefix, filters }: IFilterBarProps) {
    const [, theme] = useStyletron()
    const { q, clearQ } = useQ()
    const [t] = useTranslation()

    return (
        <div
            style={{
                display: 'flex',
                alignItems: 'center',
                borderBottom: `1px solid ${theme.borders.border200.borderColor}`,
                paddingBottom: 10,
            }}
        >
            <div>
                {Object.keys(q).length > 0 && (
                    <Button
                        startEnhancer={<AiOutlineClear size={12} />}
                        kind='tertiary'
                        size='mini'
                        onClick={() => clearQ()}
                    >
                        {t('clear search keyword, filters and sorts')}
                    </Button>
                )}
            </div>
            <div
                style={{
                    flexGrow: 1,
                }}
            />
            <div
                style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 40,
                }}
            >
                {prefix}
                <div
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 40,
                        flexShrink: 0,
                    }}
                >
                    {filters.map((filter, idx) => (
                        // eslint-disable-next-line react/jsx-props-no-spreading
                        <Filter key={idx} {...filter} />
                    ))}
                </div>
            </div>
        </div>
    )
}
