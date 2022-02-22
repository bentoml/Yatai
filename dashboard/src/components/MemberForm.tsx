import { ICreateMembersSchema, IMemberSchema } from '@/schemas/member'
import React, { useCallback, useEffect, useState } from 'react'
import { createForm } from '@/components/Form'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import UserSelector from './UserSelector'
import MemberRoleSelector from './MemberRoleSelector'

const { Form, FormItem } = createForm<ICreateMembersSchema>()

export interface IMemberFormProps {
    member?: IMemberSchema
    onSubmit: (data: ICreateMembersSchema) => Promise<void>
}

export default function MemberForm({ member, onSubmit }: IMemberFormProps) {
    const [initialValue, setInitialValue] = useState<ICreateMembersSchema>({
        usernames: member ? [member.user.name] : [],
        role: member ? member.role : 'guest',
    })

    useEffect(() => {
        if (!member) {
            return
        }
        setInitialValue({
            usernames: [member.user.name],
            role: member.role,
        })
    }, [member])

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
            <FormItem name='usernames' label={t('users')}>
                <UserSelector />
            </FormItem>
            <FormItem name='role' label={t('role')}>
                <MemberRoleSelector />
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
