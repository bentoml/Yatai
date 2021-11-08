import { ICreateOrganizationSchema, IOrganizationFullSchema } from '@/schemas/organization'
import React, { useCallback, useEffect, useState } from 'react'
import { createForm } from '@/components/Form'
import { Input } from 'baseui/input'
import { Textarea } from 'baseui/textarea'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { isModified } from '@/utils'

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
            <FormItem name={['config', 'aws', 'access_key_id']} label='AWS Access Key ID'>
                <Input />
            </FormItem>
            <FormItem name={['config', 'aws', 'secret_access_key']} label='AWS Secret Access Key'>
                <Input />
            </FormItem>
            <FormItem name={['config', 'aws', 's3', 'bucket_name']} label='S3 Bucket Name'>
                <Input />
            </FormItem>
            <FormItem name={['config', 'aws', 's3', 'region']} label='S3 Region'>
                <Input />
            </FormItem>
            <FormItem name={['config', 'aws', 'ecr', 'repository_uri']} label='ECR Repository URI'>
                <Input />
            </FormItem>
            <FormItem name={['config', 'aws', 'ecr', 'region']} label='ECR Region'>
                <Input />
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
