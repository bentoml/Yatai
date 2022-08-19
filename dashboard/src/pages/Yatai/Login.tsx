import Card from '@/components/Card'
import { createForm } from '@/components/Form'
import YataiLayout from '@/components/YataiLayout'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import useTranslation from '@/hooks/useTranslation'
import { ILoginUserSchema } from '@/schemas/user'
import { loginUser } from '@/services/user'
import { Button } from 'baseui/button'
import { Input } from 'baseui/input'
import qs from 'qs'
import React, { useCallback, useState } from 'react'
import logo from '@/assets/logo.svg'
import logoDark from '@/assets/logo-dark.svg'
import Text from '@/components/Text'
import { useLocation } from 'react-router-dom'
import { useStyletron } from 'baseui'

const { Form, FormItem } = createForm<ILoginUserSchema>()

export default function Login() {
    const currentThemeType = useCurrentThemeType()
    const [, theme] = useStyletron()
    const [t] = useTranslation()
    const location = useLocation()
    const [isLoading, setIsLoading] = useState(false)

    const handleFinish = useCallback(
        async (data: ILoginUserSchema) => {
            setIsLoading(true)
            try {
                await loginUser(data)
                const search = qs.parse(location.search, { ignoreQueryPrefix: true })
                let { redirect } = search
                if (redirect && typeof redirect === 'string') {
                    redirect = decodeURI(redirect)
                } else {
                    redirect = '/'
                }
                window.location.pathname = redirect
            } finally {
                setIsLoading(false)
            }
        },
        [location.search]
    )

    return (
        <YataiLayout
            style={{
                background: theme.colors.backgroundPrimary,
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
                    <Card
                        bodyStyle={{
                            padding: 40,
                            width: 500,
                        }}
                    >
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
                        <Form onFinish={handleFinish}>
                            <FormItem name='name_or_email' label={t('email')}>
                                <Input />
                            </FormItem>
                            <FormItem name='password' label={t('password')}>
                                <Input type='password' />
                            </FormItem>
                            <FormItem>
                                <div style={{ display: 'flex' }}>
                                    <div style={{ flexGrow: 1 }} />
                                    <Button isLoading={isLoading} size='compact'>
                                        {t('login')}
                                    </Button>
                                </div>
                            </FormItem>
                        </Form>
                    </Card>
                </div>
            </div>
        </YataiLayout>
    )
}
