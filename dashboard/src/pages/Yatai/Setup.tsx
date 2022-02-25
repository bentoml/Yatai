import React, { useCallback, useState } from 'react'
import { useStyletron } from 'baseui'
import YataiLayout from '@/components/YataiLayout'
import useTranslation from '@/hooks/useTranslation'
import { IRegisterUserSchema } from '@/schemas/user'
import { createForm } from '@/components/Form'
import Card from '@/components/Card'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import logo from '@/assets/logo.svg'
import Text from '@/components/Text'
import logoDark from '@/assets/logo-dark.svg'
import { useHistory } from 'react-router-dom'
import { Input } from 'baseui/input'
import { Button } from 'baseui/button'

const { Form, FormItem } = createForm<IRegisterUserSchema>()

export default function Setup() {
    const currentThemeType = useCurrentThemeType()
    const [, theme] = useStyletron()
    const [t] = useTranslation()
    const [isLoading, setIsLoading] = useState(false)
    const history = useHistory()
    const handleFinish = useCallback(
        async (data: IRegisterUserSchema) => {
            setIsLoading(true)
            try {
                console.log(data) // eslint-disable-line
                history.push('/')
            } finally {
                setIsLoading(false)
            }
        },
        [history]
    )

    return (
        <YataiLayout
            style={{
                background: theme.colors.background,
            }}
        >
            <div
                style={{
                    display: 'flex',
                    width: '100%',
                    height: '100%',
                    flexDirection: 'row',
                    justifyContent: 'center',
                }}
            >
                <div
                    style={{
                        display: 'flex',
                        flexDirection: 'column',
                        justifyContent: 'center',
                    }}
                >
                    <Card style={{ flexShrink: 0 }}>
                        <div
                            style={{
                                flexShrink: 0,
                                display: 'flex',
                                paddingBottom: 20,
                                alignItems: 'center',
                                gap: 10,
                            }}
                        >
                            <img
                                style={{
                                    width: 46,
                                    height: 46,
                                    display: 'inline-flex',
                                    transition: 'all 250ms cubic-bezier(0.7, 0.1, 0.33, 1) 0ms',
                                }}
                                src={currentThemeType === 'light' ? logo : logoDark}
                                alt='logo'
                            />
                            <Text
                                style={{
                                    fontSize: '18px',
                                    fontFamily: 'Zen Tokyo Zoo',
                                }}
                            >
                                YATAI
                            </Text>
                        </div>
                        <div
                            style={{
                                flexShrink: 0,
                                display: 'flex',
                                paddingBottom: 10,
                                alignItems: 'center',
                                gap: 10,
                            }}
                        >
                            Setup inital admin account
                        </div>
                        <Form onFinish={handleFinish}>
                            <FormItem name='name' label={t('name')}>
                                <Input />
                            </FormItem>
                            <FormItem name='email' label={t('email')}>
                                <Input />
                            </FormItem>
                            <FormItem name='password' label={t('password')}>
                                <Input type='password' />
                            </FormItem>
                            <FormItem>
                                <div style={{ display: 'flex' }}>
                                    <div style={{ flexGrow: 1 }}>
                                        <Button isLoading={isLoading} size='compact'>
                                            {t('submit')}
                                        </Button>
                                    </div>
                                </div>
                            </FormItem>
                        </Form>
                    </Card>
                </div>
            </div>
        </YataiLayout>
    )
}
