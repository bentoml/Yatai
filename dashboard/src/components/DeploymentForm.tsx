import { ICreateDeploymentSchema, IDeploymentSchema } from '@/schemas/deployment'
import React, { useCallback, useEffect, useState } from 'react'
import { createForm } from '@/components/Form'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { Input } from 'baseui/input'
import { Textarea } from 'baseui/textarea'
import { Accordion, Panel } from 'baseui/accordion'
import { IDeploymentSnapshotSchema } from '@/schemas/deployment_snapshot'
import { RiCpuLine } from 'react-icons/ri'
import { FaMemory } from 'react-icons/fa'
import { VscServerProcess } from 'react-icons/vsc'
import { Slider } from 'baseui/slider'
import DeploymentSnapshotTypeSelector from './DeploymentSnapshotTypeSelector'
import BentoSelector from './BentoSelector'
import BentoVersionSelector from './BentoVersionSelector'
import FormGroup from './FormGroup'
import { CPUResourceInput } from './CPUResourceInput'
import MemoryResourceInput from './MemoryResourceInput'
import DeploymentSnapshotCanaryRulesForm from './DeploymentSnapshotCanaryRulesForm'

const { Form, FormItem, useForm } = createForm<ICreateDeploymentSchema>()

export interface IDeploymentFormProps {
    orgName: string
    deployment?: IDeploymentSchema
    deploymentSnapshot?: IDeploymentSnapshotSchema
    onSubmit: (data: ICreateDeploymentSchema) => Promise<void>
}

export default function DeploymentForm({ orgName, deployment, deploymentSnapshot, onSubmit }: IDeploymentFormProps) {
    const [form] = useForm()

    const [values, setValues] = useState<ICreateDeploymentSchema>({
        name: '',
        type: 'stable',
        description: '',
        bento_name: '',
        bento_version: '',
        config: {
            hpa_conf: {
                min_replicas: 2,
                max_replicas: 10,
            },
            resources: {
                requests: {
                    cpu: '500m',
                    memory: '500Mi',
                    gpu: '',
                },
                limits: {
                    cpu: '1000m',
                    memory: '1024Mi',
                    gpu: '',
                },
            },
        },
    })

    useEffect(() => {
        form.setFieldsValue(values)
    }, [form, values])

    useEffect(() => {
        if (!deploymentSnapshot || !deployment) {
            return
        }
        setValues({
            name: deployment.name,
            description: deployment.description,
            type: deploymentSnapshot.type,
            bento_name: deploymentSnapshot.bento_version.bento.name,
            bento_version: deploymentSnapshot.bento_version.version,
            config: deploymentSnapshot.config,
        })
    }, [deployment, deploymentSnapshot])

    const [loading, setLoading] = useState(false)

    const handleFinish = useCallback(async () => {
        if (!values) {
            return
        }
        setLoading(true)
        try {
            await onSubmit({
                type: 'stable',
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                ...(values as any),
            })
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
            {!deployment && (
                <FormItem name='name' label={t('name')}>
                    <Input />
                </FormItem>
            )}
            {!deployment && (
                <FormItem name='description' label={t('description')}>
                    <Textarea />
                </FormItem>
            )}
            {deployment && (
                <FormItem name='type' label={t('type')}>
                    <DeploymentSnapshotTypeSelector />
                </FormItem>
            )}
            {values.type === 'canary' && (
                <FormItem name='canary_rules' label={t('canary rules')}>
                    <DeploymentSnapshotCanaryRulesForm />
                </FormItem>
            )}
            <FormItem name='bento_name' label={t('bento')}>
                <BentoSelector orgName={orgName} />
            </FormItem>
            {values?.bento_name && (
                <FormItem name='bento_version' label={t('bento version')}>
                    <BentoVersionSelector orgName={orgName} bentoName={values.bento_name} />
                </FormItem>
            )}
            <Accordion
                overrides={{
                    Root: {
                        style: {
                            marginBottom: '10px',
                        },
                    },
                }}
            >
                <Panel title={t('advance')}>
                    <FormGroup
                        icon={VscServerProcess}
                        style={{
                            marginTop: 30,
                        }}
                    >
                        {/* eslint-disable-next-line jsx-a11y/label-has-associated-control */}
                        <label
                            style={{
                                fontWeight: 500,
                            }}
                        >
                            {t('replicas')}
                        </label>
                        <Slider
                            min={0}
                            max={10}
                            step={1}
                            persistentThumb
                            value={[
                                values?.config?.hpa_conf?.min_replicas === undefined
                                    ? 2
                                    : values?.config?.hpa_conf?.min_replicas,
                                values?.config?.hpa_conf?.max_replicas === undefined
                                    ? 10
                                    : values?.config?.hpa_conf?.max_replicas,
                            ]}
                            onChange={({ value }) => {
                                if (!value) {
                                    return
                                }
                                setValues((values_) => {
                                    return {
                                        ...values_,
                                        config: {
                                            ...values_?.config,
                                            hpa_conf: {
                                                ...values_?.config?.hpa_conf,
                                                min_replicas: value[0],
                                                max_replicas: value[1],
                                            },
                                        },
                                    } as ICreateDeploymentSchema
                                })
                            }}
                        />
                    </FormGroup>
                    <FormGroup icon={RiCpuLine}>
                        <FormItem name={['config', 'resources', 'requests', 'cpu']} label={t('cpu requests')}>
                            <CPUResourceInput />
                        </FormItem>
                        <FormItem name={['config', 'resources', 'limits', 'cpu']} label={t('cpu limits')}>
                            <CPUResourceInput />
                        </FormItem>
                    </FormGroup>
                    <FormGroup icon={FaMemory}>
                        <FormItem name={['config', 'resources', 'requests', 'memory']} label={t('memory requests')}>
                            <MemoryResourceInput />
                        </FormItem>
                        <FormItem name={['config', 'resources', 'limits', 'memory']} label={t('memory limits')}>
                            <MemoryResourceInput />
                        </FormItem>
                    </FormGroup>
                </Panel>
            </Accordion>
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
