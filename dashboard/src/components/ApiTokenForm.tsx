import { ICreateApiTokenSchema, IApiTokenFullSchema } from '@/schemas/api_token'
import React, { useCallback, useEffect, useState } from 'react'
import { createForm } from '@/components/Form'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { Input } from 'baseui/input'
import { Textarea } from 'baseui/textarea'
import { isModified } from '@/utils'
import ApiTokenScopesCheckbox from './ApiTokenScopesCheckbox'
import DatePicker from './DatePicker'

const { Form, FormItem } = createForm<ICreateApiTokenSchema>()

export interface IApiTokenFormProps {
    apiToken?: IApiTokenFullSchema
    onSubmit: (data: ICreateApiTokenSchema) => Promise<void>
}

export default function ApiTokenForm({ apiToken, onSubmit }: IApiTokenFormProps) {
    const [values, setValues] = useState<ICreateApiTokenSchema | undefined>(apiToken)

    useEffect(() => {
        if (!apiToken) {
            return
        }
        setValues(apiToken)
    }, [apiToken])

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
            <FormItem required name='name' label={t('name')}>
                <Input disabled={apiToken !== undefined} />
            </FormItem>
            <FormItem name='description' label={t('description')}>
                <Textarea />
            </FormItem>
            <FormItem required name='scopes' label={t('scopes')}>
                <ApiTokenScopesCheckbox />
            </FormItem>
            <FormItem name='expired_at' label={t('expired_at')}>
                <DatePicker />
            </FormItem>
            <FormItem>
                <div style={{ display: 'flex' }}>
                    <div style={{ flexGrow: 1 }} />
                    <Button isLoading={loading} size={ButtonSize.compact} disabled={!isModified(apiToken, values)}>
                        {t('submit')}
                    </Button>
                </div>
            </FormItem>
        </Form>
    )
}
