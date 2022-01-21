import { ICreateClusterSchema, IClusterFullSchema } from '@/schemas/cluster'
import React, { useCallback, useEffect, useState } from 'react'
import { createForm } from '@/components/Form'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { Input } from 'baseui/input'
import { Textarea } from 'baseui/textarea'
import { isModified } from '@/utils'

const { Form, FormItem } = createForm<ICreateClusterSchema>()

export interface IClusterFormProps {
    cluster?: IClusterFullSchema
    onSubmit: (data: ICreateClusterSchema) => Promise<void>
}

export default function ClusterForm({ cluster, onSubmit }: IClusterFormProps) {
    const [values, setValues] = useState<ICreateClusterSchema | undefined>(cluster)

    useEffect(() => {
        if (!cluster) {
            return
        }
        setValues(cluster)
    }, [cluster])

    const [loading, setLoading] = useState(false)

    const handleValuesChange = useCallback((_changes, values_) => {
        setValues(values_)
    }, [])

    const handleFinish = useCallback(
        async (values_) => {
            setLoading(true)
            try {
                await onSubmit(values_)
            } finally {
                setLoading(false)
            }
        },
        [onSubmit]
    )

    const [t] = useTranslation()

    return (
        <Form initialValues={values} onFinish={handleFinish} onValuesChange={handleValuesChange}>
            <FormItem name='name' label={t('name')}>
                <Input disabled={cluster !== undefined} />
            </FormItem>
            <FormItem name='description' label={t('description')}>
                <Textarea />
            </FormItem>
            <FormItem name='kube_config' label={t('kube_config')}>
                <Textarea rows={7} />
            </FormItem>
            <FormItem name={['config', 'ingress_ip']} label='Ingress IPv4 address or hostname'>
                <Input />
            </FormItem>
            <FormItem
                name={['config', 'default_deployment_kube_namespace']}
                label={t('the default kube namespace for deployments')}
            >
                <Input />
            </FormItem>
            <FormItem>
                <div style={{ display: 'flex' }}>
                    <div style={{ flexGrow: 1 }} />
                    <Button isLoading={loading} size={ButtonSize.compact} disabled={!isModified(cluster, values)}>
                        {t('submit')}
                    </Button>
                </div>
            </FormItem>
        </Form>
    )
}
