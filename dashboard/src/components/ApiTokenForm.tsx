import { ICreateApiTokenSchema, IApiTokenSchema } from '@/schemas/api_token'
import React, { useCallback, useEffect, useState } from 'react'
import { createForm } from '@/components/Form'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { Input } from 'baseui/input'
import { Textarea } from 'baseui/textarea'
import { isModified } from '@/utils'
import { Select } from 'baseui/select'
import moment from 'moment'
import ApiTokenScopesCheckbox from '@/components/ApiTokenScopesCheckbox'
import DatePicker from '@/components/DatePicker'
import { formatDateTime } from '@/utils/datetime'

interface IExpirationDatePicker {
    value?: string
    onChange?: (value?: string) => void
    disabled?: boolean
}

function ExpirationDatePicker({ value, onChange, disabled }: IExpirationDatePicker) {
    const [selectId, setSelectId] = useState(value ? 'custom' : 'no_expiration')
    const isCustom = selectId === 'custom'
    const [t] = useTranslation()

    return (
        <div
            style={{
                display: 'flex',
                alignItems: 'center',
                gap: 10,
            }}
        >
            <div
                style={{
                    width: 200,
                }}
            >
                <Select
                    disabled={disabled}
                    clearable={false}
                    searchable={false}
                    options={[
                        {
                            id: '7 days',
                            label: t('some days', [7]),
                            getValue: () => moment().add(7, 'days').startOf('day').toDate().toISOString(),
                        },
                        {
                            id: '30 days',
                            label: t('some days', [30]),
                            getValue: () => moment().add(30, 'days').startOf('day').toDate().toISOString(),
                        },
                        {
                            id: '60 days',
                            label: t('some days', [60]),
                            getValue: () => moment().add(60, 'days').startOf('day').toDate().toISOString(),
                        },
                        {
                            id: '90 days',
                            label: t('some days', [90]),
                            getValue: () => moment().add(90, 'days').startOf('day').toDate().toISOString(),
                        },
                        { id: 'custom', label: t('custom...') },
                        { id: 'no_expiration', label: 'No expiration', getValue: () => undefined },
                    ]}
                    value={[{ id: selectId }]}
                    placeholder='Select color'
                    onChange={(params) => {
                        if (!params.option) {
                            return
                        }
                        setSelectId(params.option.id as string)
                        if (params.option.id === 'custom') {
                            return
                        }
                        onChange?.(params.option.getValue?.())
                    }}
                />
            </div>
            <div>
                {isCustom ? (
                    <DatePicker disabled={disabled} value={value} onChange={onChange} />
                ) : (
                    <span>
                        {value
                            ? t('the token will expire on sth', [formatDateTime(value)])
                            : t('the token will never expire!')}
                    </span>
                )}
            </div>
        </div>
    )
}

const { Form, FormItem } = createForm<ICreateApiTokenSchema>()

export interface IApiTokenFormProps {
    apiToken?: IApiTokenSchema
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
                <ExpirationDatePicker disabled={apiToken !== undefined} />
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
