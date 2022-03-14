import React, { useCallback, useState } from 'react'
import qs from 'qs'
import { useStyletron } from 'baseui'
import SyntaxHighlighter from 'react-syntax-highlighter'
import { docco, dark } from 'react-syntax-highlighter/dist/esm/styles/hljs'
import YataiLayout from '@/components/YataiLayout'
import useTranslation from '@/hooks/useTranslation'
import { createForm } from '@/components/Form'
import Card from '@/components/Card'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import logo from '@/assets/logo.svg'
import Text from '@/components/Text'
import logoDark from '@/assets/logo-dark.svg'
import { useHistory, useLocation } from 'react-router-dom'
import { Input } from 'baseui/input'
import { Button } from 'baseui/button'
import { setupSelfHost } from '@/services/setup'
import { toaster } from 'baseui/toast'
import { ISetupSelfHostSchema } from '@/schemas/setup'

const { Form, FormItem } = createForm<ISetupSelfHostSchema>()

export default function Setup() {
    const currentThemeType = useCurrentThemeType()
    const [, theme] = useStyletron()
    const [t] = useTranslation()
    const [isLoading, setIsLoading] = useState(false)
    const history = useHistory()
    const location = useLocation()
    const [values, setValues] = useState<ISetupSelfHostSchema | undefined>(undefined)
    const highlightTheme = currentThemeType === 'dark' ? dark : docco
    const handleFinish = useCallback(
        async (data: ISetupSelfHostSchema) => {
            setIsLoading(true)
            try {
                const search = qs.parse(location.search, { ignoreQueryPrefix: true })
                const { token } = search
                if (token && typeof token === 'string') {
                    await setupSelfHost({ ...data, token })
                } else {
                    toaster.negative('missing token in the url', { autoHideDuration: 3000 })
                    return
                }
                toaster.positive('setup success', { autoHideDuration: 3000 })
                history.push('/')
            } finally {
                setIsLoading(false)
            }
        },
        [history, location.search]
    )
    const handleValuesChange = useCallback((_changes, newValues) => {
        setValues(newValues)
    }, [])
    const getTokenCommand =
        'kubectl get pods --selector=app.kubernetes.io/name=yatai -n yatai-system ' +
        '-o jsonpath=\'{.items[0].spec.containers[0].env[?(@.name=="YATAI_INITIALIZATION_TOKEN")].value}\''

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
                    <Card
                        style={{
                            flexShrink: 0,
                        }}
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
                                minWidth: 400,
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
                            {t('setup initial admin account')}
                        </div>
                        <Form initialValues={values} onValuesChange={handleValuesChange} onFinish={handleFinish}>
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
                        <div
                            style={{
                                flexShrink: 0,
                                display: 'flex',
                                paddingTop: 30,
                                paddingBottom: 10,
                                alignItems: 'center',
                                gap: 10,
                            }}
                        >
                            {t('get the initialization token from command')}:
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
                            <SyntaxHighlighter
                                language='bash'
                                style={highlightTheme}
                                customStyle={{
                                    margin: 0,
                                }}
                            >
                                {getTokenCommand}
                            </SyntaxHighlighter>
                        </div>
                    </Card>
                </div>
            </div>
        </YataiLayout>
    )
}
