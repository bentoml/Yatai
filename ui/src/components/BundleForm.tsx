import { ICreateBundleSchema, IBundleSchema } from '@/schemas/bundle'
import React, { useCallback, useEffect, useState } from 'react'
import { createForm } from '@/components/Form'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { Input } from 'baseui/input'
import { Textarea } from 'baseui/textarea'

const { Form, FormItem } = createForm<ICreateBundleSchema>()

export interface IBundleFormProps {
    bundle?: IBundleSchema
    onSubmit: (data: ICreateBundleSchema) => Promise<void>
}

export default function BundleForm({ bundle, onSubmit }: IBundleFormProps) {
    const [initialValue, setInitialValue] = useState<ICreateBundleSchema>()

    useEffect(() => {
        if (!bundle) {
            return
        }
        setInitialValue(bundle)
    }, [bundle])

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
            <FormItem name='name' label={t('name')}>
                <Input />
            </FormItem>
            <FormItem name='description' label={t('description')}>
                <Textarea />
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
