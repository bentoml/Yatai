import { ICreateBundleVersionSchema, IBundleVersionSchema } from '@/schemas/bundle_version'
import React, { useCallback, useEffect, useState } from 'react'
import { createForm } from '@/components/Form'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { Input } from 'baseui/input'
import { Textarea } from 'baseui/textarea'

const { Form, FormItem } = createForm<ICreateBundleVersionSchema>()

export interface IBundleVersionFormProps {
    bundleVersion?: IBundleVersionSchema
    onSubmit: (data: ICreateBundleVersionSchema) => Promise<void>
}

export default function BundleVersionForm({ bundleVersion, onSubmit }: IBundleVersionFormProps) {
    const [initialValue, setInitialValue] = useState<ICreateBundleVersionSchema>()

    useEffect(() => {
        if (!bundleVersion) {
            return
        }
        setInitialValue(bundleVersion)
    }, [bundleVersion])

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
