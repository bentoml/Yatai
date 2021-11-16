import { useStyletron } from 'baseui'
import { Button } from 'baseui/button'
import { Input } from 'baseui/input'
import React, { useCallback, useState } from 'react'

export interface IDoubleCheckFormProps {
    tips: React.ReactNode
    expected: string
    onSubmit: () => Promise<void>
    buttonLabel: string
}

export default function DoubleCheckForm({ tips, expected, onSubmit, buttonLabel }: IDoubleCheckFormProps) {
    const [value, setValue] = useState('')
    const [loading, setLoading] = useState(false)
    const handleClick = useCallback(async () => {
        setLoading(true)
        try {
            await onSubmit()
        } finally {
            setLoading(false)
        }
    }, [onSubmit])
    const [, theme] = useStyletron()

    return (
        <div
            style={{
                display: 'flex',
                flexDirection: 'column',
                gap: 8,
            }}
        >
            <div>{tips}</div>
            <Input
                size='compact'
                value={value}
                onChange={(e) => {
                    setValue(e.currentTarget.value)
                }}
            />
            <Button
                overrides={{
                    BaseButton: {
                        style: {
                            backgroundColor: theme.colors.negative,
                            color: theme.colors.white,
                        },
                    },
                }}
                size='compact'
                isLoading={loading}
                disabled={value !== expected}
                onClick={handleClick}
            >
                {buttonLabel}
            </Button>
        </div>
    )
}
