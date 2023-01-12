import { ICreateDeploymentSchema, ICreateDeploymentUISchema, IDeploymentSchema } from '@/schemas/deployment'
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
import {
    DeploymentStrategy,
    ICreateDeploymentTargetSchema,
    ICreateDeploymentTargetUISchema,
    IDeploymentTargetRunnerUISchema,
} from '@/schemas/deployment_target'
import { useStyletron } from 'baseui'
import { createUseStyles } from 'react-jss'
import { IThemedStyleProps } from '@/interfaces/IThemedStyle'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import { bentomlConfigsEnvKey, resourceIconMapping, sidebarExpandedWidth, sidebarFoldedWidth } from '@/consts'
import { SidebarContext } from '@/contexts/SidebarContext'
import color from 'color'
import { LabelMedium, LabelSmall } from 'baseui/typography'
import { useHistory } from 'react-router-dom'
import { VscDebugAll, VscServerProcess, VscSymbolVariable } from 'react-icons/vsc'
import { GrResources } from 'react-icons/gr'
import { FiAlertCircle, FiInfo, FiMaximize2, FiMinimize2 } from 'react-icons/fi'
import { fetchCluster } from '@/services/cluster'
import { useQuery } from 'react-query'
import { Tabs, Tab } from 'baseui/tabs-motion'
import { IBentoWithRepositorySchema } from '@/schemas/bento'
import { StatefulTooltip } from 'baseui/tooltip'
import { Block } from 'baseui/block'
import _ from 'lodash'
import { BiCustomize } from 'react-icons/bi'
import { GiTeamUpgrade } from 'react-icons/gi'
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
import Toggle from './Toggle'
import CopyableText from './CopyableText'
import MapInput from './MapInput'
import MonacoEditor from './MonacoEditor'
import DeploymentStrategySelector from './DeploymentStrategySelector'

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

const { Form, FormItem, useForm } = createForm<ICreateDeploymentUISchema>()

const defaultTarget: ICreateDeploymentTargetUISchema = {
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
        envs: [],
        runners: {},
        enable_ingress: true,
        deployment_strategy: 'RollingUpdate',
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

    const [uiValues, setUIValues] = useState<ICreateDeploymentUISchema>({
        cluster_name: clusterName,
        name: '',
        description: '',
        targets: [_.cloneDeep(defaultTarget)],
    })

    const [values, setValues] = useState<ICreateDeploymentSchema>({
        cluster_name: clusterName,
        name: '',
        description: '',
        targets: [],
    })

    useEffect(() => {
        setValues((values_) => {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            const customizer = (value: any, srcValue: any, key: string): any => {
                if (_.endsWith(key, '_overrides')) {
                    return value
                }
                if (typeof value === 'object' || typeof srcValue === 'object') {
                    if (value === undefined || value === null) {
                        return srcValue
                    }
                    const res = _.mergeWith(value, srcValue, customizer)
                    return res
                }
                return srcValue
            }
            const mergedValues = _.mergeWith(values_, uiValues, customizer)
            return mergedValues
        })
    }, [uiValues])

    const [bento, setBento] = useState<IBentoWithRepositorySchema>()

    const clusterInfo = useQuery(`cluster:${uiValues.cluster_name}`, () =>
        uiValues.cluster_name ? fetchCluster(uiValues.cluster_name) : Promise.resolve(undefined)
    )

    const previousDeploymentRevisionUid = useRef<string>()

    useEffect(() => {
        form.setFieldsValue(uiValues)
    }, [form, uiValues])

    const [bentomlConfs, setBentomlConfs] = useState<string[]>([])

    useEffect(() => {
        if (!deploymentRevision || !deployment) {
            return
        }
        if (previousDeploymentRevisionUid.current === deploymentRevision.uid) {
            return
        }
        previousDeploymentRevisionUid.current = deploymentRevision.uid
        setBentomlConfs(
            deploymentRevision.targets.map(
                (target) => target.config?.envs?.find((env) => env.key === bentomlConfigsEnvKey)?.value ?? ''
            )
        )
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
        setUIValues(values0)
    }, [clusterName, deployment, deploymentRevision])

    const [loading, setLoading] = useState(false)

    const handleFinish = useCallback(async () => {
        if (!values) {
            return
        }
        values.targets.forEach((target, idx) => {
            if (!target.config) {
                return
            }
            if (!target.config.envs) {
                // eslint-disable-next-line no-param-reassign
                target.config.envs = []
            }
            const confItem = {
                key: bentomlConfigsEnvKey,
                value: bentomlConfs[idx] ?? '',
            }
            if (confItem.value) {
                const configsIdx = target.config.envs.findIndex((env) => env.key === bentomlConfigsEnvKey)
                if (configsIdx >= 0) {
                    // eslint-disable-next-line no-param-reassign
                    target.config.envs[configsIdx] = confItem
                } else {
                    target.config.envs.push(confItem)
                }
            } else {
                // eslint-disable-next-line no-param-reassign
                target.config.envs = target.config.envs.filter((env) => env.key !== bentomlConfigsEnvKey)
            }
            if (!target.config.runners) {
                return
            }
            Object.keys(target.config.runners).forEach((key) => {
                // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
                const runner = target.config!.runners![key]
                if (!runner.envs) {
                    runner.envs = []
                }
                if (confItem.value) {
                    const configsIdx_ = runner.envs.findIndex((env) => env.key === bentomlConfigsEnvKey)
                    if (configsIdx_ >= 0) {
                        // eslint-disable-next-line no-param-reassign
                        runner.envs[configsIdx_] = confItem
                    } else {
                        runner.envs.push(confItem)
                    }
                } else {
                    // eslint-disable-next-line no-param-reassign
                    runner.envs = runner.envs.filter((env) => env.key !== bentomlConfigsEnvKey)
                }
            })
        })
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
    }, [bentomlConfs, history, onSubmit, values])

    const handleChange = useCallback((_changes, values_) => {
        setUIValues(values_)
    }, [])

    useEffect(() => {
        setUIValues((vs) => {
            if (!clusterInfo.data) {
                return vs
            }
            const values_ = {
                ...vs,
                kube_namespace: clusterInfo.data.config?.default_deployment_kube_namespace ?? 'yatai',
            }
            return values_
        })
    }, [clusterInfo.data])

    const [t] = useTranslation()

    const [runnerTabsActiveKey, setRunnerTabsActiveKey] = useState<React.Key>()

    useEffect(() => {
        setRunnerTabsActiveKey((runnerTabsActiveKey_) => {
            if (
                bento?.manifest?.runners &&
                bento.manifest.runners.length > 0 &&
                !bento.manifest.runners.find((runner) => runner.name === runnerTabsActiveKey_)
            ) {
                return bento.manifest.runners[0].name
            }
            return runnerTabsActiveKey_
        })
        setUIValues((vs) => {
            if (deploymentRevision) {
                return vs
            }
            if (!bento?.manifest?.runners) {
                return vs
            }
            if (vs.targets.length === 0) {
                return vs
            }
            const values_ = {
                ...vs,
                targets: [
                    {
                        ...vs.targets[0],
                        config: {
                            ...vs.targets[0].config,
                            runners:
                                bento?.manifest?.runners?.reduce((runners, runner) => {
                                    const conf = {
                                        // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
                                        resources: _.cloneDeep(defaultTarget.config!.resources!),
                                        // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
                                        hpa_conf: _.cloneDeep(defaultTarget.config!.hpa_conf!),
                                        deployment_strategy: 'RollingUpdate' as DeploymentStrategy,
                                    }
                                    if (runner.resource_config?.cpu) {
                                        if (!conf.resources.requests) {
                                            conf.resources.requests = {
                                                cpu: String(runner.resource_config.cpu),
                                                memory: '500Mi',
                                                gpu: '',
                                            }
                                        } else {
                                            conf.resources.requests.cpu = String(runner.resource_config.cpu)
                                        }
                                        if (!conf.resources.limits) {
                                            conf.resources.limits = {
                                                cpu: String(runner.resource_config.cpu),
                                                memory: '500Mi',
                                                gpu: '',
                                            }
                                        } else {
                                            conf.resources.limits.cpu = String(runner.resource_config.cpu)
                                        }
                                    }
                                    return {
                                        ...runners,
                                        [runner.name]: conf,
                                    }
                                }, {} as Record<string, IDeploymentTargetRunnerUISchema>) ?? {},
                        },
                    },
                ],
            }
            return values_
        })
    }, [bento?.manifest?.runners, deploymentRevision])

    const [showJsonEditor, setShowJsonEditor] = useState(false)

    return (
        <div className={styles.wrapper}>
            <div
                className={styles.header}
                style={{
                    left: (ctx.expanded ? sidebarExpandedWidth : sidebarFoldedWidth) + 1,
                }}
            >
                <div className={styles.headerLabel}>
                    <LabelMedium>{deployment ? t('update deployment') : t('new deployment')}</LabelMedium>
                </div>
                <div
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 20,
                    }}
                >
                    <Toggle
                        labelPlacement='right'
                        value={showJsonEditor}
                        onChange={() => setShowJsonEditor((v_) => !v_)}
                    >
                        JSON Editor
                    </Toggle>
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
                    <Button isLoading={loading} size={ButtonSize.compact} onClick={handleFinish}>
                        {t('submit')}
                    </Button>
                </div>
            </div>
            <div className={styles.body}>
                {showJsonEditor ? (
                    <MonacoEditor
                        value={JSON.stringify(values, null, 4)}
                        height='calc(100vh - 170px)'
                        theme={themeType === 'dark' ? 'vs-dark' : 'Dawn'}
                        defaultLanguage='json'
                        options={{
                            minimap: {
                                enabled: true,
                            },
                            lineNumbers: 'off',
                            lineDecorationsWidth: 0,
                        }}
                        onChange={(value) => {
                            if (!value) {
                                return
                            }
                            const v = JSON.parse(value)
                            setValues(v as ICreateDeploymentSchema)
                            setUIValues(v as ICreateDeploymentUISchema)
                        }}
                    />
                ) : (
                    <Form form={form} initialValues={uiValues} onValuesChange={handleChange}>
                        <div>
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
                                {uiValues.cluster_name && (
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
                                {uiValues.targets.map((target, idx) => {
                                    return (
                                        <div key={idx}>
                                            <FormItem
                                                required
                                                name={['targets', idx, 'config', 'enable_ingress']}
                                                label={t('endpoint public access')}
                                            >
                                                <Toggle labelPlacement='right'>
                                                    <div
                                                        style={{
                                                            display: 'flex',
                                                            alignItems: 'center',
                                                            gap: 3,
                                                        }}
                                                    >
                                                        <span style={{ fontSize: '12px', fontWeight: 'normal' }}>
                                                            {target?.config?.enable_ingress
                                                                ? t('enabled')
                                                                : t('disabled')}
                                                        </span>
                                                        {target?.config?.enable_ingress ? (
                                                            <StatefulTooltip
                                                                showArrow
                                                                content={() => (
                                                                    <Block width={['100px', '200px', '400px', '600px']}>
                                                                        <span>
                                                                            {t('endpoint public access enable piece 1')}{' '}
                                                                        </span>
                                                                        <span style={{ fontWeight: 'bold' }}>
                                                                            {t('warning')}:{' '}
                                                                        </span>
                                                                        <span>
                                                                            {t('endpoint public access enable piece 2')}
                                                                        </span>
                                                                    </Block>
                                                                )}
                                                            >
                                                                <div>
                                                                    <FiAlertCircle size={12} />
                                                                </div>
                                                            </StatefulTooltip>
                                                        ) : (
                                                            <StatefulTooltip
                                                                showArrow
                                                                content={() => (
                                                                    <Block width={['100px', '200px', '400px', '600px']}>
                                                                        <span>
                                                                            {t(
                                                                                'endpoint public access disable piece 1'
                                                                            )}{' '}
                                                                        </span>
                                                                        <span>
                                                                            <CopyableText
                                                                                highlight
                                                                                text='kubectl port-forward'
                                                                            />{' '}
                                                                        </span>
                                                                        <span>
                                                                            {t(
                                                                                'endpoint public access disable piece 2'
                                                                            )}
                                                                        </span>
                                                                    </Block>
                                                                )}
                                                            >
                                                                <div>
                                                                    <FiInfo size={12} />
                                                                </div>
                                                            </StatefulTooltip>
                                                        )}
                                                    </div>
                                                </Toggle>
                                            </FormItem>
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
                                                                {React.createElement(
                                                                    resourceIconMapping.bento_repository,
                                                                    {}
                                                                )}
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
                                                            <BentoSelector
                                                                bentoRepositoryName={target.bento_repository}
                                                                onBentoChange={setBento}
                                                            />
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
                                                        <LabelSmall>{t('bentoml configuration')}</LabelSmall>
                                                    </div>
                                                    <div
                                                        style={{
                                                            paddingLeft: paddingLeft + 5,
                                                            marginTop: 10,
                                                            width: 400,
                                                        }}
                                                    >
                                                        <MonacoEditor
                                                            value={bentomlConfs[idx] ?? ''}
                                                            height='200px'
                                                            theme={themeType === 'dark' ? 'vs-dark' : 'Dawn'}
                                                            defaultLanguage='yaml'
                                                            options={{
                                                                minimap: {
                                                                    enabled: false,
                                                                },
                                                                lineNumbers: 'off',
                                                                lineDecorationsWidth: 0,
                                                            }}
                                                            onChange={(value) => {
                                                                setBentomlConfs((bentomlConfs_) => {
                                                                    const newBentomlConfs = bentomlConfs_.slice()
                                                                    // eslint-disable-next-line @typescript-eslint/no-explicit-any
                                                                    newBentomlConfs[idx] = value ?? ''
                                                                    return newBentomlConfs
                                                                })
                                                            }}
                                                        />
                                                    </div>
                                                    <div
                                                        style={{
                                                            display: 'flex',
                                                            alignItems: 'center',
                                                            gap: 10,
                                                            marginTop: 20,
                                                        }}
                                                    >
                                                        <VscServerProcess />
                                                        <LabelSmall>{t('number of replicas')}</LabelSmall>
                                                    </div>
                                                    <div
                                                        style={{
                                                            paddingLeft,
                                                            marginTop: 10,
                                                        }}
                                                    >
                                                        <FormItem
                                                            name={[
                                                                'targets',
                                                                idx,
                                                                'config',
                                                                'hpa_conf',
                                                                'min_replicas',
                                                            ]}
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
                                                            name={[
                                                                'targets',
                                                                idx,
                                                                'config',
                                                                'hpa_conf',
                                                                'max_replicas',
                                                            ]}
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
                                                        <LabelSmall>{t('resource per replicas')}</LabelSmall>
                                                    </div>
                                                    <div
                                                        style={{
                                                            paddingLeft,
                                                            marginTop: 10,
                                                        }}
                                                    >
                                                        <FormGroup icon={RiCpuLine}>
                                                            <FormItem
                                                                name={[
                                                                    'targets',
                                                                    idx,
                                                                    'config',
                                                                    'resources',
                                                                    'requests',
                                                                    'cpu',
                                                                ]}
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
                                                                name={[
                                                                    'targets',
                                                                    idx,
                                                                    'config',
                                                                    'resources',
                                                                    'limits',
                                                                    'cpu',
                                                                ]}
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
                                                                name={[
                                                                    'targets',
                                                                    idx,
                                                                    'config',
                                                                    'resources',
                                                                    'requests',
                                                                    'memory',
                                                                ]}
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
                                                                name={[
                                                                    'targets',
                                                                    idx,
                                                                    'config',
                                                                    'resources',
                                                                    'limits',
                                                                    'memory',
                                                                ]}
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
                                                        <FormGroup icon={BiCustomize}>
                                                            <FormItem
                                                                name={[
                                                                    'targets',
                                                                    idx,
                                                                    'config',
                                                                    'resources',
                                                                    'requests',
                                                                    'custom',
                                                                ]}
                                                                label={t('custom resources requests')}
                                                            >
                                                                <MapInput
                                                                    style={{
                                                                        width: '600px',
                                                                    }}
                                                                />
                                                            </FormItem>
                                                            <FormItem
                                                                name={[
                                                                    'targets',
                                                                    idx,
                                                                    'config',
                                                                    'resources',
                                                                    'limits',
                                                                    'custom',
                                                                ]}
                                                                label={t('custom resources limits')}
                                                            >
                                                                <MapInput
                                                                    style={{
                                                                        width: '600px',
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
                                                        <FormGroup icon={GiTeamUpgrade}>
                                                            <FormItem
                                                                name={['targets', idx, 'config', 'deployment_strategy']}
                                                                label={
                                                                    <div
                                                                        style={{
                                                                            display: 'flex',
                                                                            alignItems: 'center',
                                                                            gap: 6,
                                                                        }}
                                                                    >
                                                                        <div>{t('deployment strategy')}</div>
                                                                        <StatefulTooltip
                                                                            showArrow
                                                                            content={() => (
                                                                                <Block
                                                                                    width={[
                                                                                        '100px',
                                                                                        '200px',
                                                                                        '400px',
                                                                                        '600px',
                                                                                    ]}
                                                                                >
                                                                                    <p>
                                                                                        {t('deployment strategy desc')}
                                                                                    </p>
                                                                                    <p>{t('RollingUpdate desc')}</p>
                                                                                    <p>{t('Recreate desc')}</p>
                                                                                    <p>{t('RampedSlowRollout desc')}</p>
                                                                                    <p>
                                                                                        {t(
                                                                                            'BestEffortControlledRollout desc'
                                                                                        )}
                                                                                    </p>
                                                                                </Block>
                                                                            )}
                                                                        >
                                                                            <div>
                                                                                <FiAlertCircle size={12} />
                                                                            </div>
                                                                        </StatefulTooltip>
                                                                    </div>
                                                                }
                                                                style={{ width: 340 }}
                                                            >
                                                                <DeploymentStrategySelector />
                                                            </FormItem>
                                                        </FormGroup>
                                                        <FormGroup icon={VscDebugAll}>
                                                            <FormItem
                                                                name={[
                                                                    'targets',
                                                                    idx,
                                                                    'config',
                                                                    'enable_stealing_traffic_debug_mode',
                                                                ]}
                                                                label={t('enable stealing traffic debug mode')}
                                                            >
                                                                <Toggle labelPlacement='right'>
                                                                    <div
                                                                        style={{
                                                                            display: 'flex',
                                                                            alignItems: 'center',
                                                                            gap: 3,
                                                                        }}
                                                                    >
                                                                        <span
                                                                            style={{
                                                                                fontSize: '12px',
                                                                                fontWeight: 'normal',
                                                                            }}
                                                                        >
                                                                            {target?.config
                                                                                ?.enable_stealing_traffic_debug_mode
                                                                                ? t('enabled')
                                                                                : t('disabled')}
                                                                        </span>
                                                                        <StatefulTooltip
                                                                            showArrow
                                                                            content={() => (
                                                                                <Block
                                                                                    width={[
                                                                                        '100px',
                                                                                        '200px',
                                                                                        '400px',
                                                                                        '600px',
                                                                                    ]}
                                                                                >
                                                                                    <span>
                                                                                        {t(
                                                                                            'enable stealing traffic debug mode desc'
                                                                                        )}
                                                                                    </span>
                                                                                </Block>
                                                                            )}
                                                                        >
                                                                            <div>
                                                                                <FiAlertCircle size={12} />
                                                                            </div>
                                                                        </StatefulTooltip>
                                                                    </div>
                                                                </Toggle>
                                                            </FormItem>
                                                            <FormItem
                                                                name={[
                                                                    'targets',
                                                                    idx,
                                                                    'config',
                                                                    'enable_debug_pod_receive_production_traffic',
                                                                ]}
                                                                label={t('enable debug pod receive production traffic')}
                                                            >
                                                                <Toggle labelPlacement='right'>
                                                                    <div
                                                                        style={{
                                                                            display: 'flex',
                                                                            alignItems: 'center',
                                                                            gap: 3,
                                                                        }}
                                                                    >
                                                                        <span
                                                                            style={{
                                                                                fontSize: '12px',
                                                                                fontWeight: 'normal',
                                                                            }}
                                                                        >
                                                                            {target?.config
                                                                                ?.enable_debug_pod_receive_production_traffic
                                                                                ? t('enabled')
                                                                                : t('disabled')}
                                                                        </span>
                                                                        <StatefulTooltip
                                                                            showArrow
                                                                            content={() => (
                                                                                <Block
                                                                                    width={[
                                                                                        '100px',
                                                                                        '200px',
                                                                                        '400px',
                                                                                        '600px',
                                                                                    ]}
                                                                                >
                                                                                    <span>
                                                                                        {t(
                                                                                            'enable debug pod receive production traffic desc'
                                                                                        )}
                                                                                    </span>
                                                                                </Block>
                                                                            )}
                                                                        >
                                                                            <div>
                                                                                <FiAlertCircle size={12} />
                                                                            </div>
                                                                        </StatefulTooltip>
                                                                    </div>
                                                                </Toggle>
                                                            </FormItem>
                                                        </FormGroup>
                                                        <FormGroup icon={VscSymbolVariable}>
                                                            <FormItem
                                                                name={['targets', idx, 'config', 'envs']}
                                                                label={t('environment variables')}
                                                            >
                                                                <LabelList
                                                                    ignoreKeys={[bentomlConfigsEnvKey]}
                                                                    style={{
                                                                        width: 440,
                                                                    }}
                                                                />
                                                            </FormItem>
                                                        </FormGroup>
                                                    </Panel>
                                                </Accordion>
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
                                                    <Panel title='Runners'>
                                                        <Tabs
                                                            activeKey={runnerTabsActiveKey}
                                                            renderAll
                                                            onChange={({ activeKey }) => {
                                                                setRunnerTabsActiveKey(activeKey)
                                                                setUIValues((values_) => {
                                                                    const runner =
                                                                        values_.targets[0]?.config?.runners?.[activeKey]
                                                                    if (runner) {
                                                                        return values_
                                                                    }
                                                                    const vs = {
                                                                        ...values_,
                                                                        targets: [
                                                                            {
                                                                                ...values_.targets[0],
                                                                                config: {
                                                                                    ...values_.targets[0].config,
                                                                                    runners: {
                                                                                        ...values_.targets[0]?.config
                                                                                            ?.runners,
                                                                                        [activeKey]: {},
                                                                                    },
                                                                                },
                                                                            },
                                                                        ],
                                                                    }
                                                                    return vs
                                                                })
                                                            }}
                                                        >
                                                            {bento?.manifest.runners?.map((runner) => (
                                                                <Tab
                                                                    key={runner.name}
                                                                    title={
                                                                        <div
                                                                            style={{
                                                                                display: 'flex',
                                                                                alignItems: 'center',
                                                                                gap: 8,
                                                                            }}
                                                                        >
                                                                            {React.createElement(
                                                                                resourceIconMapping.bento_runner,
                                                                                {
                                                                                    size: 12,
                                                                                }
                                                                            )}
                                                                            <span>{runner.name}</span>
                                                                        </div>
                                                                    }
                                                                >
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
                                                                            <LabelSmall>
                                                                                {t('number of replicas')}
                                                                            </LabelSmall>
                                                                        </div>
                                                                        <div
                                                                            style={{
                                                                                paddingLeft,
                                                                                marginTop: 10,
                                                                            }}
                                                                        >
                                                                            <FormItem
                                                                                name={[
                                                                                    'targets',
                                                                                    idx,
                                                                                    'config',
                                                                                    'runners',
                                                                                    runner.name,
                                                                                    'hpa_conf',
                                                                                    'min_replicas',
                                                                                ]}
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
                                                                                name={[
                                                                                    'targets',
                                                                                    idx,
                                                                                    'config',
                                                                                    'runners',
                                                                                    runner.name,
                                                                                    'hpa_conf',
                                                                                    'max_replicas',
                                                                                ]}
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
                                                                            <LabelSmall>
                                                                                {t('resource per replicas')}
                                                                            </LabelSmall>
                                                                        </div>
                                                                        <div
                                                                            style={{
                                                                                paddingLeft,
                                                                                marginTop: 10,
                                                                            }}
                                                                        >
                                                                            <FormGroup icon={RiCpuLine}>
                                                                                <FormItem
                                                                                    name={[
                                                                                        'targets',
                                                                                        idx,
                                                                                        'config',
                                                                                        'runners',
                                                                                        runner.name,
                                                                                        'resources',
                                                                                        'requests',
                                                                                        'cpu',
                                                                                    ]}
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
                                                                                    name={[
                                                                                        'targets',
                                                                                        idx,
                                                                                        'config',
                                                                                        'runners',
                                                                                        runner.name,
                                                                                        'resources',
                                                                                        'limits',
                                                                                        'cpu',
                                                                                    ]}
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
                                                                                    name={[
                                                                                        'targets',
                                                                                        idx,
                                                                                        'config',
                                                                                        'runners',
                                                                                        runner.name,
                                                                                        'resources',
                                                                                        'requests',
                                                                                        'memory',
                                                                                    ]}
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
                                                                                    name={[
                                                                                        'targets',
                                                                                        idx,
                                                                                        'config',
                                                                                        'runners',
                                                                                        runner.name,
                                                                                        'resources',
                                                                                        'limits',
                                                                                        'memory',
                                                                                    ]}
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
                                                                            <FormGroup icon={BiCustomize}>
                                                                                <FormItem
                                                                                    name={[
                                                                                        'targets',
                                                                                        idx,
                                                                                        'config',
                                                                                        'runners',
                                                                                        runner.name,
                                                                                        'resources',
                                                                                        'requests',
                                                                                        'custom',
                                                                                    ]}
                                                                                    label={t(
                                                                                        'custom resources requests'
                                                                                    )}
                                                                                >
                                                                                    <MapInput
                                                                                        style={{
                                                                                            width: '600px',
                                                                                        }}
                                                                                    />
                                                                                </FormItem>
                                                                                <FormItem
                                                                                    name={[
                                                                                        'targets',
                                                                                        idx,
                                                                                        'config',
                                                                                        'runners',
                                                                                        runner.name,
                                                                                        'resources',
                                                                                        'limits',
                                                                                        'custom',
                                                                                    ]}
                                                                                    label={t('custom resources limits')}
                                                                                >
                                                                                    <MapInput
                                                                                        style={{
                                                                                            width: '600px',
                                                                                        }}
                                                                                    />
                                                                                </FormItem>
                                                                            </FormGroup>
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
                                                                                <FormGroup icon={GiTeamUpgrade}>
                                                                                    <FormItem
                                                                                        name={[
                                                                                            'targets',
                                                                                            idx,
                                                                                            'config',
                                                                                            'runners',
                                                                                            runner.name,
                                                                                            'deployment_strategy',
                                                                                        ]}
                                                                                        label={
                                                                                            <div
                                                                                                style={{
                                                                                                    display: 'flex',
                                                                                                    alignItems:
                                                                                                        'center',
                                                                                                    gap: 6,
                                                                                                }}
                                                                                            >
                                                                                                <div>
                                                                                                    {t(
                                                                                                        'deployment strategy'
                                                                                                    )}
                                                                                                </div>
                                                                                                <StatefulTooltip
                                                                                                    showArrow
                                                                                                    content={() => (
                                                                                                        <Block
                                                                                                            width={[
                                                                                                                '100px',
                                                                                                                '200px',
                                                                                                                '400px',
                                                                                                                '600px',
                                                                                                            ]}
                                                                                                        >
                                                                                                            <p>
                                                                                                                {t(
                                                                                                                    'deployment strategy desc'
                                                                                                                )}
                                                                                                            </p>
                                                                                                            <p>
                                                                                                                {t(
                                                                                                                    'RollingUpdate desc'
                                                                                                                )}
                                                                                                            </p>
                                                                                                            <p>
                                                                                                                {t(
                                                                                                                    'Recreate desc'
                                                                                                                )}
                                                                                                            </p>
                                                                                                            <p>
                                                                                                                {t(
                                                                                                                    'RampedSlowRollout desc'
                                                                                                                )}
                                                                                                            </p>
                                                                                                            <p>
                                                                                                                {t(
                                                                                                                    'BestEffortControlledRollout desc'
                                                                                                                )}
                                                                                                            </p>
                                                                                                        </Block>
                                                                                                    )}
                                                                                                >
                                                                                                    <div>
                                                                                                        <FiAlertCircle
                                                                                                            size={12}
                                                                                                        />
                                                                                                    </div>
                                                                                                </StatefulTooltip>
                                                                                            </div>
                                                                                        }
                                                                                        style={{ width: 340 }}
                                                                                    >
                                                                                        <DeploymentStrategySelector />
                                                                                    </FormItem>
                                                                                </FormGroup>
                                                                                <FormGroup icon={VscDebugAll}>
                                                                                    <FormItem
                                                                                        name={[
                                                                                            'targets',
                                                                                            idx,
                                                                                            'config',
                                                                                            'runners',
                                                                                            runner.name,
                                                                                            'enable_stealing_traffic_debug_mode',
                                                                                        ]}
                                                                                        label={t(
                                                                                            'enable stealing traffic debug mode'
                                                                                        )}
                                                                                    >
                                                                                        <Toggle labelPlacement='right'>
                                                                                            <div
                                                                                                style={{
                                                                                                    display: 'flex',
                                                                                                    alignItems:
                                                                                                        'center',
                                                                                                    gap: 3,
                                                                                                }}
                                                                                            >
                                                                                                <span
                                                                                                    style={{
                                                                                                        fontSize:
                                                                                                            '12px',
                                                                                                        fontWeight:
                                                                                                            'normal',
                                                                                                    }}
                                                                                                >
                                                                                                    {target?.config
                                                                                                        ?.runners?.[
                                                                                                        runner.name
                                                                                                    ]
                                                                                                        ?.enable_stealing_traffic_debug_mode
                                                                                                        ? t('enabled')
                                                                                                        : t('disabled')}
                                                                                                </span>
                                                                                                <StatefulTooltip
                                                                                                    showArrow
                                                                                                    content={() => (
                                                                                                        <Block
                                                                                                            width={[
                                                                                                                '100px',
                                                                                                                '200px',
                                                                                                                '400px',
                                                                                                                '600px',
                                                                                                            ]}
                                                                                                        >
                                                                                                            <span>
                                                                                                                {t(
                                                                                                                    'enable stealing traffic debug mode desc'
                                                                                                                )}
                                                                                                            </span>
                                                                                                        </Block>
                                                                                                    )}
                                                                                                >
                                                                                                    <div>
                                                                                                        <FiAlertCircle
                                                                                                            size={12}
                                                                                                        />
                                                                                                    </div>
                                                                                                </StatefulTooltip>
                                                                                            </div>
                                                                                        </Toggle>
                                                                                    </FormItem>
                                                                                    <FormItem
                                                                                        name={[
                                                                                            'targets',
                                                                                            idx,
                                                                                            'config',
                                                                                            'runners',
                                                                                            runner.name,
                                                                                            'enable_debug_pod_receive_production_traffic',
                                                                                        ]}
                                                                                        label={t(
                                                                                            'enable debug pod receive production traffic'
                                                                                        )}
                                                                                    >
                                                                                        <Toggle labelPlacement='right'>
                                                                                            <div
                                                                                                style={{
                                                                                                    display: 'flex',
                                                                                                    alignItems:
                                                                                                        'center',
                                                                                                    gap: 3,
                                                                                                }}
                                                                                            >
                                                                                                <span
                                                                                                    style={{
                                                                                                        fontSize:
                                                                                                            '12px',
                                                                                                        fontWeight:
                                                                                                            'normal',
                                                                                                    }}
                                                                                                >
                                                                                                    {target?.config
                                                                                                        ?.runners?.[
                                                                                                        runner.name
                                                                                                    ]
                                                                                                        ?.enable_debug_pod_receive_production_traffic
                                                                                                        ? t('enabled')
                                                                                                        : t('disabled')}
                                                                                                </span>
                                                                                                <StatefulTooltip
                                                                                                    showArrow
                                                                                                    content={() => (
                                                                                                        <Block
                                                                                                            width={[
                                                                                                                '100px',
                                                                                                                '200px',
                                                                                                                '400px',
                                                                                                                '600px',
                                                                                                            ]}
                                                                                                        >
                                                                                                            <span>
                                                                                                                {t(
                                                                                                                    'enable debug pod receive production traffic desc'
                                                                                                                )}
                                                                                                            </span>
                                                                                                        </Block>
                                                                                                    )}
                                                                                                >
                                                                                                    <div>
                                                                                                        <FiAlertCircle
                                                                                                            size={12}
                                                                                                        />
                                                                                                    </div>
                                                                                                </StatefulTooltip>
                                                                                            </div>
                                                                                        </Toggle>
                                                                                    </FormItem>
                                                                                </FormGroup>
                                                                                <FormGroup icon={VscSymbolVariable}>
                                                                                    <FormItem
                                                                                        name={[
                                                                                            'targets',
                                                                                            idx,
                                                                                            'config',
                                                                                            'runners',
                                                                                            runner.name,
                                                                                            'envs',
                                                                                        ]}
                                                                                        label={t(
                                                                                            'environment variables'
                                                                                        )}
                                                                                    >
                                                                                        <LabelList
                                                                                            ignoreKeys={[
                                                                                                bentomlConfigsEnvKey,
                                                                                            ]}
                                                                                            style={{
                                                                                                width: 440,
                                                                                            }}
                                                                                        />
                                                                                    </FormItem>
                                                                                </FormGroup>
                                                                            </Panel>
                                                                        </Accordion>
                                                                    </div>
                                                                </Tab>
                                                            ))}
                                                        </Tabs>
                                                    </Panel>
                                                </Accordion>
                                            </div>
                                        </div>
                                    )
                                })}
                            </div>
                        </div>
                    </Form>
                )}
            </div>
        </div>
    )
}
