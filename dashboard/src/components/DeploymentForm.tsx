/* eslint-disable @typescript-eslint/ban-ts-comment */
// @ts-nocheck
import { ICreateDeploymentSchema, IDeploymentSchema } from '@/schemas/deployment'
import { DeleteAlt } from 'baseui/icon'
import React, { useCallback, useEffect, useState } from 'react'
import { createForm } from '@/components/Form'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { Input } from 'baseui/input'
import { Textarea } from 'baseui/textarea'
import { Accordion, Panel } from 'baseui/accordion'
import { IDeploymentRevisionSchema } from '@/schemas/deployment_revision'
import { RiCpuLine } from 'react-icons/ri'
import { FaMemory } from 'react-icons/fa'
import { VscServerProcess } from 'react-icons/vsc'
import { Slider } from 'baseui/slider'
import { ICreateDeploymentTargetSchema } from '@/schemas/deployment_target'
import { useStyletron } from 'baseui'
import DeploymentTargetTypeSelector from './DeploymentTargetTypeSelector'
import BentoSelector from './BentoSelector'
import BentoVersionSelector from './BentoVersionSelector'
import FormGroup from './FormGroup'
import { CPUResourceInput } from './CPUResourceInput'
import MemoryResourceInput from './MemoryResourceInput'
import DeploymentTargetCanaryRulesForm from './DeploymentTargetCanaryRulesForm'
import ClusterSelector from './ClusterSelector'
import Label from './Label'

const { Form, FormList, FormItem, useForm } = createForm<ICreateDeploymentSchema>()

const defaultTarget: ICreateDeploymentTargetSchema = {
    type: 'stable',
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
}

export interface IDeploymentFormProps {
    clusterName?: string
    deployment?: IDeploymentSchema
    deploymentRevision?: IDeploymentRevisionSchema
    onSubmit: (data: ICreateDeploymentSchema) => Promise<void>
}

export default function DeploymentForm({
    clusterName,
    deployment,
    deploymentRevision,
    onSubmit,
}: IDeploymentFormProps) {
    const [form] = useForm()

    const [, theme] = useStyletron()

    const [values, setValues] = useState<ICreateDeploymentSchema>({
        cluster_name: clusterName,
        name: '',
        description: '',
        targets: [defaultTarget],
    })

    useEffect(() => {
        form.setFieldsValue(values)
    }, [form, values])

    useEffect(() => {
        if (!deploymentRevision || !deployment) {
            return
        }
        setValues({
            name: deployment.name,
            description: deployment.description,
            cluster_name: clusterName,
            targets: deploymentRevision.targets.map(
                (target) =>
                    ({
                        type: target.type,
                        bento_name: target.bento_version.bento.name,
                        bento_version: target.bento_version.version,
                        canary_rules: target.canary_rules,
                        config: target.config,
                    } as ICreateDeploymentTargetSchema)
            ),
        })
    }, [clusterName, deployment, deploymentRevision])

    const [loading, setLoading] = useState(false)

    const addTarget = useCallback(() => {
        setValues((values_) => {
            return {
                ...values_,
                targets: [
                    ...values_.targets,
                    {
                        ...defaultTarget,
                        type: values_.targets.length > 0 ? 'canary' : 'stable',
                    },
                ],
            }
        })
    }, [])

    const removeTarget = useCallback((idx: number) => {
        setValues((values_) => {
            return {
                ...values_,
                targets: values_.targets.filter((_target, idx_) => idx !== idx_),
            }
        })
    }, [])

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
            <FormItem
                required
                name='cluster_name'
                label={t('cluster')}
                style={{ display: clusterName ? 'none' : 'block' }}
            >
                <ClusterSelector />
            </FormItem>
            {!deployment && (
                <FormItem required name='name' label={t('name')}>
                    <Input />
                </FormItem>
            )}
            {!deployment && (
                <FormItem name='description' label={t('description')}>
                    <Textarea />
                </FormItem>
            )}
            <Label style={{ paddingBottom: 10, display: 'block' }}>{t('sth list', [t('deployment target')])} *</Label>
            <FormList name='targets'>
                {(fields) => (
                    <div
                        style={{
                            background: theme.colors.backgroundSecondary,
                            marginBottom: 10,
                        }}
                    >
                        {fields
                            .filter(({ name }) => values.targets[name] !== undefined)
                            .map(({ key, name }) => {
                                const target = values.targets[name]
                                return (
                                    <div
                                        key={key}
                                        style={{
                                            borderBottom: `1px solid ${theme.borders.border400.borderColor}`,
                                            padding: '10px 10px 10px 20px',
                                        }}
                                    >
                                        <div>
                                            <FormItem required name={[name, 'type']} label={t('type')}>
                                                <DeploymentTargetTypeSelector />
                                            </FormItem>
                                            {target.type === 'canary' && (
                                                <FormItem
                                                    required
                                                    name={[name, 'canary_rules']}
                                                    label={t('canary rules')}
                                                >
                                                    <DeploymentTargetCanaryRulesForm
                                                        style={{
                                                            paddingLeft: 40,
                                                        }}
                                                    />
                                                </FormItem>
                                            )}
                                            <FormItem required name={[name, 'bento_name']} label={t('bento')}>
                                                <BentoSelector />
                                            </FormItem>
                                            {target.bento_name && (
                                                <FormItem
                                                    required
                                                    name={[name, 'bento_version']}
                                                    label={t('bento version')}
                                                >
                                                    <BentoVersionSelector bentoName={target.bento_name} />
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
                                                <Panel title={t('advanced')}>
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
                                                                target.config?.hpa_conf?.min_replicas === undefined
                                                                    ? 2
                                                                    : target.config?.hpa_conf?.min_replicas,
                                                                target.config?.hpa_conf?.max_replicas === undefined
                                                                    ? 10
                                                                    : target.config?.hpa_conf?.max_replicas,
                                                            ]}
                                                            onChange={({ value }) => {
                                                                if (!value) {
                                                                    return
                                                                }
                                                                setValues((values_) => {
                                                                    return {
                                                                        ...values_,
                                                                        config: {
                                                                            ...target.config,
                                                                            hpa_conf: {
                                                                                ...target.config?.hpa_conf,
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
                                                        <FormItem
                                                            name={[name, 'config', 'resources', 'requests', 'cpu']}
                                                            label={t('cpu requests')}
                                                        >
                                                            <CPUResourceInput />
                                                        </FormItem>
                                                        <FormItem
                                                            name={[name, 'config', 'resources', 'limits', 'cpu']}
                                                            label={t('cpu limits')}
                                                        >
                                                            <CPUResourceInput />
                                                        </FormItem>
                                                    </FormGroup>
                                                    <FormGroup icon={FaMemory}>
                                                        <FormItem
                                                            name={[name, 'config', 'resources', 'requests', 'memory']}
                                                            label={t('memory requests')}
                                                        >
                                                            <MemoryResourceInput />
                                                        </FormItem>
                                                        <FormItem
                                                            name={[name, 'config', 'resources', 'limits', 'memory']}
                                                            label={t('memory limits')}
                                                        >
                                                            <MemoryResourceInput />
                                                        </FormItem>
                                                    </FormGroup>
                                                </Panel>
                                            </Accordion>
                                        </div>
                                        <Button
                                            size='mini'
                                            disabled={values.targets.length === 1}
                                            overrides={{
                                                Root: {
                                                    style: {
                                                        background: theme.colors.negative,
                                                    },
                                                },
                                            }}
                                            onClick={(e) => {
                                                e.preventDefault()
                                                removeTarget(name)
                                            }}
                                        >
                                            <DeleteAlt />
                                            <span style={{ marginLeft: 6 }}>{t('delete')}</span>
                                        </Button>
                                    </div>
                                )
                            })}
                        <div style={{ padding: 10 }}>
                            <Button
                                size='mini'
                                onClick={(e) => {
                                    e.preventDefault()
                                    addTarget()
                                }}
                            >
                                {t('add sth', [t('deployment target')])}
                            </Button>
                        </div>
                    </div>
                )}
            </FormList>

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
