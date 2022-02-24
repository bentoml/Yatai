import React, { useCallback, useEffect, useState } from 'react' // eslint-disable-line
import { createForm } from '@/components/Form'
import { IChangePasswordUISchema } from '@/schemas/user'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import useTranslation from '@/hooks/useTranslation'
import { Input } from 'baseui/input'

const { Form, FormItem } = createForm<IChangePasswordUISchema>()

export interface IChangePasswordFormProps {
    onSubmit: (data: IChangePasswordUISchema) => Promise<void>
}

export default function PasswordForm({ onSubmit }: IChangePasswordFormProps) {
    const [initialValue] = useState<IChangePasswordUISchema | undefined>(undefined)
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
            <FormItem name='current_password' label='Current password'>
                <Input type='password' />
            </FormItem>
            <FormItem name='new_password' label='New password'>
                <Input type='password' />
            </FormItem>
            <FormItem name='confirm_new_password' label='Confirm new password'>
                <Input type='password' />
            </FormItem>
            <FormItem>
                <div style={{ display: 'flex' }}>
                    <div style={{ flex: 1 }} />
                    <Button isLoading={loading} type='submit' size={ButtonSize.compact}>
                        {t('submit')}
                    </Button>
                </div>
            </FormItem>
        </Form>
    )
}
