import { ICreateBentoSchema, IBentoSchema } from '@/schemas/bento'
import React, { useCallback, useEffect, useState } from 'react'
import { createForm } from '@/components/Form'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { Input } from 'baseui/input'
import { Textarea } from 'baseui/textarea'

const { Form, FormItem } = createForm<ICreateBentoSchema>()

export interface IBentoFormProps {
    bento?: IBentoSchema
    onSubmit: (data: ICreateBentoSchema) => Promise<void>
}

export default function BentoForm({ bento, onSubmit }: IBentoFormProps) {
    const [initialValue, setInitialValue] = useState<ICreateBentoSchema>()

    useEffect(() => {
        if (!bento) {
            return
        }
        setInitialValue(bento)
    }, [bento])

    const [loading, setLoading] = useState(false)

    const handleFinish = useCallback(
        async (values) => {
            setLoading(true)
            try {
                await onSubmit(values)
            } finally {
                setLoading(false)
            }
        },
        [onSubmit]
    )

    const [t] = useTranslation()

    return (
        <Form initialValues={initialValue} onFinish={handleFinish}>
            <FormItem name='version' label={t('version')}>
                <Input />
            </FormItem>
            <FormItem name='description' label={t('description')}>
                <Textarea />
            </FormItem>
            <FormItem name='build_at' label={t('build_at')}>
                <Input />
            </FormItem>
            <FormItem>
                <div style={{ display: 'flex' }}>
                    <div style={{ flexGrow: 1 }} />
                    <Button isLoading={loading} size={ButtonSize.compact}>
                        {t('submit')}
                    </Button>
                </div>
            </FormItem>
        </Form>
    )
}
