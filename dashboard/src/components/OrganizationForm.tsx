import { ICreateOrganizationSchema, IOrganizationFullSchema } from '@/schemas/organization'
import React, { useCallback, useEffect, useState } from 'react'
import { createForm } from '@/components/Form'
import { Input } from 'baseui/input'
import { Textarea } from 'baseui/textarea'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { isModified } from '@/utils'
import Toggle from './Toggle'

const { Form, FormItem } = createForm<ICreateOrganizationSchema>()

export interface IOrganizationFormProps {
    organization?: IOrganizationFullSchema
    onSubmit: (data: ICreateOrganizationSchema) => Promise<void>
}

export default function OrganizationForm({ organization, onSubmit }: IOrganizationFormProps) {
    const [values, setValues] = useState<ICreateOrganizationSchema | undefined>(organization)

    useEffect(() => {
        if (!organization) {
            return
        }
        setValues({
            name: organization.name,
            description: organization.description,
            config: organization.config,
        })
    }, [organization])

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
                <Input disabled={organization !== undefined} />
            </FormItem>
            <FormItem name='description' label={t('description')}>
                <Textarea />
            </FormItem>
            <FormItem name={['config', 's3', 'endpoint']} label='S3 Endpoint'>
                <Input />
            </FormItem>
            <FormItem name={['config', 's3', 'access_key']} label='S3 Access Key'>
                <Input />
            </FormItem>
            <FormItem name={['config', 's3', 'secret_key']} label='S3 Secret Key'>
                <Input />
            </FormItem>
            <FormItem name={['config', 's3', 'bentos_bucket_name']} label='S3 Bentos Bucket Name'>
                <Input />
            </FormItem>
            <FormItem name={['config', 's3', 'models_bucket_name']} label='S3 Models Bucket Name'>
                <Input />
            </FormItem>
            <FormItem name={['config', 's3', 'region']} label='S3 Region'>
                <Input />
            </FormItem>
            <FormItem name={['config', 's3', 'secure']} label='S3 Secure'>
                <Toggle />
            </FormItem>
            <FormItem>
                <div style={{ display: 'flex' }}>
                    <div style={{ flexGrow: 1 }} />
                    <Button isLoading={loading} size={ButtonSize.compact} disabled={!isModified(organization, values)}>
                        {t('submit')}
                    </Button>
                </div>
            </FormItem>
        </Form>
    )
}
