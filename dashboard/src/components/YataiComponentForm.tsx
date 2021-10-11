import { ICreateYataiComponentSchema, IYataiComponentSchema, YataiComponentType } from '@/schemas/yatai_component'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { createForm } from '@/components/Form'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { useFetchYataiComponents } from '@/hooks/useFetchYataiComponents'
import { yataiComponentIconMapping } from '@/consts'
import YataiComponentTypeSelector from './YataiComponentTypeSelector'

const { Form, FormItem, useForm } = createForm<ICreateYataiComponentSchema>()

export interface IYataiComponentFormProps {
    orgName: string
    clusterName: string
    yataiComponent?: IYataiComponentSchema
    onSubmit: (data: ICreateYataiComponentSchema) => Promise<void>
}

export default function YataiComponentForm({
    orgName,
    clusterName,
    yataiComponent,
    onSubmit,
}: IYataiComponentFormProps) {
    const { yataiComponentsInfo } = useFetchYataiComponents(orgName, clusterName)

    const [form] = useForm()

    const [values, setValues] = useState<ICreateYataiComponentSchema>({
        type: 'logging',
    })

    useEffect(() => {
        form.setFieldsValue(values)
    }, [form, values])

    const installedTypes = useMemo(() => yataiComponentsInfo.data?.map((x) => x.type) ?? [], [yataiComponentsInfo.data])

    useEffect(() => {
        const type = Object.keys(yataiComponentIconMapping).find(
            (x) => installedTypes.indexOf(x as YataiComponentType) < 0
        )
        if (!type) {
            return
        }
        setValues((values_) => ({
            ...values_,
            type: type as YataiComponentType,
        }))
    }, [installedTypes])

    useEffect(() => {
        if (!yataiComponent) {
            return
        }
        setValues({
            type: yataiComponent.type,
        })
    }, [clusterName, yataiComponent])

    const [loading, setLoading] = useState(false)

    const handleFinish = useCallback(async () => {
        if (!values) {
            return
        }
        setLoading(true)
        try {
            await onSubmit(values)
        } finally {
            setLoading(false)
        }
    }, [onSubmit, values])

    const handleChange = useCallback((changes, values_) => {
        setValues(values_)
    }, [])

    const [t] = useTranslation()

    return (
        <Form form={form} initialValues={values} onFinish={handleFinish} onValuesChange={handleChange}>
            <FormItem required name='type' label={t('type')}>
                <YataiComponentTypeSelector excludes={installedTypes} />
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
