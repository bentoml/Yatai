import { ICreateBentoRepositorySchema, IBentoRepositorySchema } from '@/schemas/bento_repository'
import React, { useCallback, useEffect, useState } from 'react'
import { createForm } from '@/components/Form'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { Input } from 'baseui/input'
import { Textarea } from 'baseui/textarea'

const { Form, FormItem } = createForm<ICreateBentoRepositorySchema>()

export interface IBentoRepositoryFormProps {
    bentoRepository?: IBentoRepositorySchema
    onSubmit: (data: ICreateBentoRepositorySchema) => Promise<void>
}

export default function BentoRepositoryForm({ bentoRepository, onSubmit }: IBentoRepositoryFormProps) {
    const [initialValue, setInitialValue] = useState<ICreateBentoRepositorySchema>()

    useEffect(() => {
        if (!bentoRepository) {
            return
        }
        setInitialValue(bentoRepository)
    }, [bentoRepository])

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
