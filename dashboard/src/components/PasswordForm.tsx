import React, { useCallback, useState } from 'react' // eslint-disable-line
import { toaster } from 'baseui/toast'
import { createForm } from '@/components/Form'
import { IChangePasswordSchema } from '@/schemas/user'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import useTranslation from '@/hooks/useTranslation'
import { Input } from 'baseui/input'

interface IChangePasswordUISchema extends IChangePasswordSchema {
    confirm_new_password: string
}

const { Form, FormItem } = createForm<IChangePasswordUISchema>()

export interface IChangePasswordFormProps {
    onSubmit: (data: IChangePasswordSchema) => Promise<void>
}

export default function PasswordForm({ onSubmit }: IChangePasswordFormProps) {
    const [initialValue] = useState<IChangePasswordUISchema | undefined>(undefined)
    const [loading, setLoading] = useState(false)
    const [t] = useTranslation()
    const handleFinish = useCallback(
        async (values) => {
            if (values.new_password !== values.confirm_new_password) {
                toaster.negative(t('password not match'), { autoHideDuration: 3000 })
            } else {
                setLoading(true)
                try {
                    await onSubmit({
                        new_password: values.new_password,
                        current_password: values.current_password,
                    })
                } finally {
                    setLoading(false)
                }
            }
        },
        [t, onSubmit]
    )
    return (
        <Form initialValues={initialValue} onFinish={handleFinish}>
            <FormItem name='current_password' label={t('current password')}>
                <Input type='password' />
            </FormItem>
            <FormItem name='new_password' label={t('new password')}>
                <Input type='password' />
            </FormItem>
            <FormItem name='confirm_new_password' label={t('confirm password')}>
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
