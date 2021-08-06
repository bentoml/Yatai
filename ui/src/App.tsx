import React from 'react'
import { Client as Styletron } from 'styletron-engine-atomic'
import { Provider as StyletronProvider } from 'styletron-react'
import { LightTheme, BaseProvider } from 'baseui'
import { Input } from 'baseui/input'
import { Button } from 'baseui/button'
import Header from '@/components/Header'
import Layout from '@/components/Layout'
import { createForm } from '@/components/Form'

const engine = new Styletron()

interface IData {
    name: string
    age: number
}

const { Form, FormItem } = createForm<IData>()

export default function Hello() {
    return (
        <StyletronProvider value={engine}>
            <BaseProvider theme={LightTheme}>
                <Header />
                <Layout>
                    <Form
                        onFinish={(values) => {
                            console.log(values)
                        }}
                    >
                        <FormItem
                            name='age'
                            label='age'
                            required
                            validators={[
                                async (_, value?: string) => {
                                    if (!value || value.length < 3) {
                                        throw Error('less than 3 characters')
                                    }
                                },
                            ]}
                        >
                            <Input />
                        </FormItem>
                        <FormItem>
                            <Button>Submit</Button>
                        </FormItem>
                    </Form>
                </Layout>
            </BaseProvider>
        </StyletronProvider>
    )
}
