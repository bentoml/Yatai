import { ICreateUserSchema } from '@/schemas/user'
import React, { useCallback, useEffect, useState } from 'react' // eslint-disable-line
import { createForm } from '@/components/Form'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { Input } from 'baseui/input'
import MemberRoleSelector from './MemberRoleSelector'

const { Form, FormItem } = createForm<ICreateUserSchema>()

export interface ICreateUserSchemaProps {
    onSubmit: (data: ICreateUserSchema) => Promise<void>
}

export default function UserForm({ onSubmit }: ICreateUserSchemaProps) {
    const [initialValue, setInitialValue] = useState<ICreateUserSchema>({  // eslint-disable-line
        name: '',
        email: '',
        password: '',
        role: 'guest',
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
            <FormItem required name='name' label={t('name')}>
                <Input />
            </FormItem>
            <FormItem name='role' label={t('role')}>
                <MemberRoleSelector />
            </FormItem>
            <FormItem required name='email' label={t('email')}>
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
