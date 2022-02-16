import { IRegisterUserSchema } from '@/schemas/user'
import React, { useCallback, useEffect, useState } from 'react' // eslint-disable-line
import { createForm } from '@/components/Form'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { Input } from 'baseui/input'

const { Form, FormItem } = createForm<IRegisterUserSchema>()

export interface IRegisterUserFormProps {
    onSubmit: (data: IRegisterUserSchema) => Promise<void>
}

export default function RegisterUserForm({ onSubmit }: IRegisterUserFormProps) {
    const [initialValue, setInitialValue] = useState<IRegisterUserSchema>({  // eslint-disable-line
        name: '',
        first_name: '',
        last_name: '',
        email: '',
        password: '',
    })

    const [loading, setLoading] = useState(false)

    const handleFinish = useCallback(
        async (value) => {
            setLoading(true)
            try {
                await onSubmit(value)
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
            <FormItem name='first_name' label={t('first_name')}>
                <Input />
            </FormItem>
            <FormItem name='last_name' label={t('last_name')}>
                <Input />
            </FormItem>
            <FormItem name='email' label={t('email')}>
                <Input />
            </FormItem>
            <FormItem name='password' label={t('password')}>
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
