import { ICreateClusterSchema, IClusterFullSchema } from '@/schemas/cluster'
import React, { useCallback, useEffect, useState } from 'react'
import { createForm } from '@/components/Form'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { Input } from 'baseui/input'
import { Textarea } from 'baseui/textarea'

const { Form, FormItem } = createForm<ICreateClusterSchema>()

export interface IClusterFormProps {
    cluster?: IClusterFullSchema
    onSubmit: (data: ICreateClusterSchema) => Promise<void>
}

export default function ClusterForm({ cluster, onSubmit }: IClusterFormProps) {
    const [initialValue, setInitialValue] = useState<ICreateClusterSchema | undefined>()

    useEffect(() => {
        if (!cluster) {
            return
        }
        setInitialValue(cluster)
    }, [cluster])

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
            <FormItem name='kube_config' label={t('kube_config')}>
                <Textarea />
            </FormItem>
            <FormItem name={['config', 'ingress_ip']} label='Ingress IP'>
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
