import { ICreateOrganizationSchema, IOrganizationSchema } from '@/schemas/organization'
import React, { useCallback, useEffect, useState } from 'react'
import { createForm } from '@/components/Form'
import { Input } from 'baseui/input'
import { Textarea } from 'baseui/textarea'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'

const { Form, FormItem } = createForm<ICreateOrganizationSchema>()

export interface IOrganizationFormProps {
    organization?: IOrganizationSchema
    onSubmit: (data: ICreateOrganizationSchema) => Promise<void>
}

export default function OrganizationForm({ organization, onSubmit }: IOrganizationFormProps) {
    const [initialValue, setInitialValue] = useState<ICreateOrganizationSchema | undefined>()

    useEffect(() => {
        if (!organization) {
            return
        }
        setInitialValue({
            name: organization.name,
            description: organization.description,
        })
    }, [organization])

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
