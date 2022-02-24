import React, { useCallback, useEffect, useState } from 'react' // eslint-disable-line
import { createForm } from '@/components/Form'
import { IChangePasswordSchema } from '@/schemas/user'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import useTranslation from '@/hooks/useTranslation'
import { Input } from 'baseui/input'

const { Form, FormItem } = createForm<IChangePasswordSchema>()

export interface IChangePasswordFormProps {
    onSubmit: (data: IChangePasswordSchema) => Promise<void>
}

export default function PasswordForm({ onSubmit }: IChangePasswordFormProps) {
    const [values, setValues] = useState<IChangePasswordSchema | undefined>(undefined)
    const handleFinish = useCallback(
        async (values_) => {
            try {
                await onSubmit(values_)
            } finally {
                console.log('something') // eslint-disable-line
            }
        },
        [onSubmit]
    )
    const [t] = useTranslation()
    return (
        <Form initialValues={values} onFinish={handleFinish}>
            <FormItem name='current_password' label='Current password'>
                <Input type='password' />
            </FormItem>
            <FormItem name='new_password' label='New password'>
                <Input type='password' />
            </FormItem>
            <FormItem>
                <div style={{ display: 'flex' }}>
                    <div style={{ flex: 1 }} />
                    <Button type='submit' size={ButtonSize.compact}>
                        {t('submit')}
                    </Button>
                </div>
            </FormItem>
        </Form>
    )
}
