import { ICreateDeploymentSchema, IDeploymentSchema } from '@/schemas/deployment'
import React, { useCallback, useContext, useEffect, useRef, useState } from 'react'
import { createForm } from '@/components/Form'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { Input } from 'baseui/input'
import { Textarea } from 'baseui/textarea'
import { Accordion, Panel } from 'baseui/accordion'
import { IDeploymentRevisionSchema } from '@/schemas/deployment_revision'
import { RiCpuLine } from 'react-icons/ri'
import { FaMemory } from 'react-icons/fa'
import { ICreateDeploymentTargetSchema } from '@/schemas/deployment_target'
import { useStyletron } from 'baseui'
import { createUseStyles } from 'react-jss'
import { IThemedStyleProps } from '@/interfaces/IThemedStyle'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import { resourceIconMapping, sidebarExpandedWidth, sidebarFoldedWidth } from '@/consts'
import { SidebarContext } from '@/contexts/SidebarContext'
import color from 'color'
import { Label2, Label3 } from 'baseui/typography'
import { useHistory } from 'react-router-dom'
import { VscServerProcess, VscSymbolVariable } from 'react-icons/vsc'
import { GrResources } from 'react-icons/gr'
import { FiMaximize2, FiMinimize2 } from 'react-icons/fi'
import { fetchCluster } from '@/services/cluster'
import { useQuery } from 'react-query'
import DeploymentTargetTypeSelector from './DeploymentTargetTypeSelector'
import BentoRepositorySelector from './BentoRepositorySelector'
import BentoSelector from './BentoSelector'
import FormGroup from './FormGroup'
import { CPUResourceInput } from './CPUResourceInput'
import MemoryResourceInput from './MemoryResourceInput'
import DeploymentTargetCanaryRulesForm from './DeploymentTargetCanaryRulesForm'
import ClusterSelector from './ClusterSelector'
import Divider from './Divider'
import LabelList from './LabelList'
import NumberInput from './NumberInput'

const useStyles = createUseStyles({
    wrapper: () => {
        return {
            width: '100%',
            paddingBottom: '40px',
        }
    },
    header: (props: IThemedStyleProps) => {
        return {
            boxSizing: 'border-box',
            display: 'flex',
            alignItems: 'center',
            borderBottom: `1px solid ${props.theme.borders.border300.borderColor}`,
            background: color(props.theme.colors.backgroundPrimary).fade(0.5).rgb().string(),
            backdropFilter: 'blur(10px)',
            padding: '0.9rem 1rem',
            marginTop: '-20px',
            right: 0,
            left: 0,
            position: 'fixed',
            zIndex: 999,
            transition: 'all 200ms cubic-bezier(0.7, 0.1, 0.33, 1) 0ms',
        }
    },
    body: {
        paddingTop: '70px',
    },
    headerLabel: {
        flexGrow: 1,
    },
})

const { Form, FormItem, useForm } = createForm<ICreateDeploymentSchema>()

const defaultTarget: ICreateDeploymentTargetSchema = {
    type: 'stable',
    bento_repository: '',
    bento: '',
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
    const paddingLeft = 20
    const ctx = useContext(SidebarContext)
    const themeType = useCurrentThemeType()
    const [, theme] = useStyletron()
    const styles = useStyles({ theme, themeType })
    const [form] = useForm()
    const history = useHistory()

    const [values, setValues] = useState<ICreateDeploymentSchema>({
        cluster_name: clusterName,
        name: '',
        description: '',
        targets: [defaultTarget],
    })

    const clusterInfo = useQuery(`cluster:${values.cluster_name}`, () =>
        values.cluster_name ? fetchCluster(values.cluster_name) : Promise.resolve(undefined)
    )

    const previousDeploymentRevisionUid = useRef<string>()

    useEffect(() => {
        form.setFieldsValue(values)
    }, [form, values])

    useEffect(() => {
        if (!deploymentRevision || !deployment) {
            return
        }
        if (previousDeploymentRevisionUid.current === deploymentRevision.uid) {
            return
        }
        previousDeploymentRevisionUid.current = deploymentRevision.uid
        const values0 = {
            name: deployment.name,
            description: deployment.description,
            cluster_name: clusterName || deployment.cluster?.name,
            kube_namespace: deployment.kube_namespace,
            targets: deploymentRevision.targets.map((target) => {
                return {
                    type: target.type,
                    bento_repository: target.bento.repository.name,
                    bento: target.bento.version,
                    canary_rules: target.canary_rules,
                    config: target.config,
                } as ICreateDeploymentTargetSchema
            }),
        }
        setValues(values0)
    }, [clusterName, deployment, deploymentRevision])

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
            history.goBack()
        } finally {
            setLoading(false)
        }
    }, [history, onSubmit, values])

    const handleChange = useCallback((_changes, values_) => {
        setValues(values_)
    }, [])

    useEffect(() => {
        setValues((vs) => {
            if (!clusterInfo.data) {
                return vs
            }
            return {
                ...vs,
                kube_namespace: clusterInfo.data.config?.default_deployment_kube_namespace ?? 'yatai',
            }
        })
    }, [clusterInfo.data])

    const [t] = useTranslation()

    return (
        <Form
            className={styles.wrapper}
            form={form}
            initialValues={values}
            onFinish={handleFinish}
            onValuesChange={handleChange}
        >
            <div
                className={styles.header}
                style={{
                    left: (ctx.expanded ? sidebarExpandedWidth : sidebarFoldedWidth) + 1,
                }}
            >
                <div className={styles.headerLabel}>
                    <Label2>{deployment ? t('update deployment') : t('new deployment')}</Label2>
                </div>
                <div
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 20,
                    }}
                >
                    <Button
                        isLoading={loading}
                        size={ButtonSize.compact}
                        kind='secondary'
                        onClick={(e) => {
                            e.preventDefault()
                            history.goBack()
                        }}
                    >
                        {t('cancel')}
                    </Button>
                    <Button isLoading={loading} size={ButtonSize.compact}>
                        {t('submit')}
                    </Button>
                </div>
            </div>

            <div className={styles.body}>
                <div
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 20,
                    }}
                >
                    <FormItem
                        required
                        name='cluster_name'
                        label={t('cluster')}
                        style={{ display: clusterName ? 'none' : 'block', marginBottom: 0 }}
                    >
                        <ClusterSelector
                            disabled={!!deployment}
                            overrides={{
                                Root: {
                                    style: {
                                        width: '392px',
                                    },
                                },
                            }}
                        />
                    </FormItem>
                    {values.cluster_name && (
                        <FormItem
                            required
                            name='kube_namespace'
                            label={t('kube namespace')}
                            style={{ marginBottom: 0 }}
                        >
                            <Input disabled={!!deployment} />
                        </FormItem>
                    )}
                </div>
                <FormItem required name='name' label={t('deployment name')}>
                    <Input
                        disabled={!!deployment}
                        overrides={{
                            Root: {
                                style: {
                                    width: '392px',
                                },
                            },
                        }}
                    />
                </FormItem>
                {!deployment && (
                    <FormItem
                        name='description'
                        label={t('description')}
                        style={{
                            width: 838,
                        }}
                    >
                        <Textarea />
                    </FormItem>
                )}
                <div>
                    {values.targets.map((target, idx) => {
                        return (
                            <div key={idx}>
                                <Divider orientation='left'>{t('select bento')}</Divider>
                                <div>
                                    <FormItem
                                        style={{ display: 'none' }}
                                        required
                                        name={['targets', idx, 'type']}
                                        label={t('type')}
                                    >
                                        <DeploymentTargetTypeSelector />
                                    </FormItem>
                                    {target.type === 'canary' && (
                                        <FormItem
                                            required
                                            name={['targets', idx, 'canary_rules']}
                                            label={t('canary rules')}
                                        >
                                            <DeploymentTargetCanaryRulesForm
                                                style={{
                                                    paddingLeft: 40,
                                                }}
                                            />
                                        </FormItem>
                                    )}
                                    <div
                                        style={{
                                            paddingLeft,
                                            display: 'flex',
                                            alignItems: 'center',
                                            gap: 20,
                                        }}
                                    >
                                        <FormItem
                                            required
                                            name={['targets', idx, 'bento_repository']}
                                            style={{ marginBottom: 0, width: 370 }}
                                            label={
                                                <div
                                                    style={{
                                                        display: 'flex',
                                                        alignItems: 'center',
                                                        gap: 5,
                                                    }}
                                                >
                                                    {React.createElement(resourceIconMapping.bento_repository, {})}
                                                    <div>{t('bento repository')}</div>
                                                </div>
                                            }
                                        >
                                            <BentoRepositorySelector />
                                        </FormItem>
                                        {target.bento_repository && (
                                            <FormItem
                                                required
                                                name={['targets', idx, 'bento']}
                                                label={
                                                    <div
                                                        style={{
                                                            display: 'flex',
                                                            alignItems: 'center',
                                                            gap: 5,
                                                        }}
                                                    >
                                                        {React.createElement(resourceIconMapping.bento, {})}
                                                        <div>{t('bento')}</div>
                                                    </div>
                                                }
                                                style={{ width: 370 }}
                                            >
                                                <BentoSelector bentoRepositoryName={target.bento_repository} />
                                            </FormItem>
                                        )}
                                    </div>
                                    <Divider orientation='left'>{t('configurations')}</Divider>
                                    <div
                                        style={{
                                            paddingLeft,
                                        }}
                                    >
                                        <div
                                            style={{
                                                display: 'flex',
                                                alignItems: 'center',
                                                gap: 10,
                                            }}
                                        >
                                            <VscServerProcess />
                                            <Label3>{t('number of replicas')}</Label3>
                                        </div>
                                        <div
                                            style={{
                                                paddingLeft,
                                                marginTop: 10,
                                            }}
                                        >
                                            <FormItem
                                                name={['targets', idx, 'config', 'hpa_conf', 'min_replicas']}
                                                label={
                                                    <div
                                                        style={{
                                                            display: 'flex',
                                                            alignItems: 'center',
                                                            gap: 16,
                                                        }}
                                                    >
                                                        <FiMinimize2 size={20} />
                                                        <div>{t('min')}</div>
                                                    </div>
                                                }
                                            >
                                                <NumberInput
                                                    overrides={{
                                                        Root: {
                                                            style: {
                                                                width: '220px',
                                                                marginLeft: '36px',
                                                            },
                                                        },
                                                    }}
                                                />
                                            </FormItem>
                                            <FormItem
                                                name={['targets', idx, 'config', 'hpa_conf', 'max_replicas']}
                                                label={
                                                    <div
                                                        style={{
                                                            display: 'flex',
                                                            alignItems: 'center',
                                                            gap: 16,
                                                        }}
                                                    >
                                                        <FiMaximize2 size={20} />
                                                        <div>{t('max')}</div>
                                                    </div>
                                                }
                                            >
                                                <NumberInput
                                                    overrides={{
                                                        Root: {
                                                            style: {
                                                                width: '220px',
                                                                marginLeft: '36px',
                                                            },
                                                        },
                                                    }}
                                                />
                                            </FormItem>
                                        </div>
                                        <div
                                            style={{
                                                display: 'flex',
                                                alignItems: 'center',
                                                gap: 10,
                                            }}
                                        >
                                            <GrResources />
                                            <Label3>{t('resource per replicas')}</Label3>
                                        </div>
                                        <div
                                            style={{
                                                paddingLeft,
                                                marginTop: 10,
                                            }}
                                        >
                                            <FormGroup icon={RiCpuLine}>
                                                <FormItem
                                                    name={['targets', idx, 'config', 'resources', 'requests', 'cpu']}
                                                    label={t('cpu requests')}
                                                >
                                                    <CPUResourceInput
                                                        overrides={{
                                                            Root: {
                                                                style: {
                                                                    width: '220px',
                                                                },
                                                            },
                                                        }}
                                                    />
                                                </FormItem>
                                                <FormItem
                                                    name={['targets', idx, 'config', 'resources', 'limits', 'cpu']}
                                                    label={t('cpu limits')}
                                                >
                                                    <CPUResourceInput
                                                        overrides={{
                                                            Root: {
                                                                style: {
                                                                    width: '220px',
                                                                },
                                                            },
                                                        }}
                                                    />
                                                </FormItem>
                                            </FormGroup>
                                            <FormGroup icon={FaMemory}>
                                                <FormItem
                                                    name={['targets', idx, 'config', 'resources', 'requests', 'memory']}
                                                    label={t('memory requests')}
                                                >
                                                    <MemoryResourceInput
                                                        overrides={{
                                                            Root: {
                                                                style: {
                                                                    width: '130px',
                                                                },
                                                            },
                                                        }}
                                                    />
                                                </FormItem>
                                                <FormItem
                                                    name={['targets', idx, 'config', 'resources', 'limits', 'memory']}
                                                    label={t('memory limits')}
                                                >
                                                    <MemoryResourceInput
                                                        overrides={{
                                                            Root: {
                                                                style: {
                                                                    width: '130px',
                                                                },
                                                            },
                                                        }}
                                                    />
                                                </FormItem>
                                            </FormGroup>
                                        </div>
                                    </div>
                                    <Accordion
                                        overrides={{
                                            Root: {
                                                style: {
                                                    marginBottom: '10px',
                                                },
                                            },
                                        }}
                                        renderAll
                                    >
                                        <Panel title={t('advanced')}>
                                            <FormItem
                                                name={['targets', idx, 'config', 'envs']}
                                                label={
                                                    <div
                                                        style={{
                                                            display: 'flex',
                                                            alignItems: 'center',
                                                            gap: 5,
                                                        }}
                                                    >
                                                        <VscSymbolVariable />
                                                        <div>{t('environment variables')}</div>
                                                    </div>
                                                }
                                            >
                                                <LabelList
                                                    style={{
                                                        width: 440,
                                                    }}
                                                />
                                            </FormItem>
                                        </Panel>
                                    </Accordion>
                                </div>
                            </div>
                        )
                    })}
                </div>
            </div>
        </Form>
    )
}
