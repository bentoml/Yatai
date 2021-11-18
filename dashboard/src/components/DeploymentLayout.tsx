import { resourceIconMapping } from '@/consts'
import { useCluster } from '@/hooks/useCluster'
import { useDeployment, useDeploymentLoading } from '@/hooks/useDeployment'
import { useFetchDeploymentRevisions } from '@/hooks/useFetchDeploymentRevisions'
import { useFetchYataiComponents } from '@/hooks/useFetchYataiComponents'
import { useOrganization } from '@/hooks/useOrganization'
import { usePage } from '@/hooks/usePage'
import { useSubscription } from '@/hooks/useSubscription'
import useTranslation from '@/hooks/useTranslation'
import { IDeploymentFullSchema, IDeploymentSchema, IUpdateDeploymentSchema } from '@/schemas/deployment'
import { deleteDeployment, fetchDeployment, terminateDeployment, updateDeployment } from '@/services/deployment'
import { useStyletron } from 'baseui'
import { Button } from 'baseui/button'
import { Modal, ModalBody, ModalHeader } from 'baseui/modal'
import color from 'color'
import _ from 'lodash'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { AiOutlineDashboard } from 'react-icons/ai'
import { FaJournalWhills } from 'react-icons/fa'
import { RiSurveyLine } from 'react-icons/ri'
import { VscServerProcess } from 'react-icons/vsc'
import { useQuery, useQueryClient } from 'react-query'
import { useHistory, useParams } from 'react-router-dom'
import { INavItem } from './BaseSidebar'
import BaseSubLayout from './BaseSubLayout'
import Card from './Card'
import DeploymentForm from './DeploymentForm'
import DeploymentStatusTag from './DeploymentStatusTag'
import DoubleCheckForm from './DoubleCheckForm'

export interface IDeploymentLayoutProps {
    children: React.ReactNode
}

export default function DeploymentLayout({ children }: IDeploymentLayoutProps) {
    const { clusterName, deploymentName } = useParams<{ clusterName: string; deploymentName: string }>()
    const queryKey = `fetchDeployment:${clusterName}:${deploymentName}`
    const deploymentInfo = useQuery(queryKey, () => fetchDeployment(clusterName, deploymentName))
    const { deployment, setDeployment } = useDeployment()
    const { organization, setOrganization } = useOrganization()
    const { cluster, setCluster } = useCluster()
    const { setDeploymentLoading } = useDeploymentLoading()
    useEffect(() => {
        setDeploymentLoading(deploymentInfo.isLoading)
        if (deploymentInfo.isSuccess) {
            if (!_.isEqual(deployment, deploymentInfo.data)) {
                setDeployment(deploymentInfo.data)
            }
            if (deploymentInfo.data.cluster?.uid !== cluster?.uid) {
                setCluster(deploymentInfo.data.cluster)
            }
            if (deploymentInfo.data.cluster?.organization?.uid !== organization?.uid) {
                setOrganization(deploymentInfo.data.cluster?.organization)
            }
        } else if (deploymentInfo.isLoading) {
            setDeployment(undefined)
        }
    }, [
        cluster?.uid,
        deployment,
        deployment?.uid,
        deploymentInfo.data,
        deploymentInfo.isLoading,
        deploymentInfo.isSuccess,
        organization?.uid,
        setCluster,
        setDeployment,
        setDeploymentLoading,
        setOrganization,
    ])

    const uids = useMemo(() => (deploymentInfo.data?.uid ? [deploymentInfo.data.uid] : []), [deploymentInfo.data?.uid])
    const queryClient = useQueryClient()
    const subscribeCb = useCallback(
        (deployment_: IDeploymentSchema) => {
            queryClient.setQueryData(queryKey, (oldData?: IDeploymentFullSchema): IDeploymentFullSchema => {
                if (oldData && oldData.uid !== deployment_.uid) {
                    return oldData
                }
                return { ...oldData, ...deployment_ }
            })
        },
        [queryClient, queryKey]
    )
    const { subscribe, unsubscribe } = useSubscription()

    useEffect(() => {
        subscribe({
            resourceType: 'deployment',
            resourceUids: uids,
            cb: subscribeCb,
        })
        return () => {
            unsubscribe({
                resourceType: 'deployment',
                resourceUids: uids,
                cb: subscribeCb,
            })
        }
    }, [subscribe, subscribeCb, uids, unsubscribe])

    const [t] = useTranslation()

    const { yataiComponentsInfo } = useFetchYataiComponents(clusterName)
    const hasLogging = yataiComponentsInfo.data?.find((x) => x.type === 'logging') !== undefined
    const hasMonitoring = yataiComponentsInfo.data?.find((x) => x.type === 'monitoring') !== undefined

    const [page] = usePage()
    const { deploymentRevisionsInfo } = useFetchDeploymentRevisions(clusterName, deploymentName, page)
    const [isCreateDeploymentRevisionOpen, setIsCreateDeploymentRevisionOpen] = useState(false)
    const handleCreateDeploymentRevision = useCallback(
        async (data: IUpdateDeploymentSchema) => {
            await updateDeployment(clusterName, deploymentName, data)
            await deploymentInfo.refetch()
            await deploymentRevisionsInfo.refetch()
            setIsCreateDeploymentRevisionOpen(false)
        },
        [clusterName, deploymentName, deploymentInfo, deploymentRevisionsInfo]
    )

    const breadcrumbItems: INavItem[] = useMemo(
        () => [
            {
                title: t('sth list', [t('cluster')]),
                path: '/clusters',
                icon: resourceIconMapping.cluster,
            },
            {
                title: clusterName,
                path: `/clusters/${clusterName}`,
            },
            {
                title: t('sth list', [t('deployment')]),
                path: `/clusters/${clusterName}/deployments`,
                icon: resourceIconMapping.deployment,
            },
            {
                title: deploymentName,
                path: `/clusters/${clusterName}/deployments/${deploymentName}`,
            },
        ],
        [clusterName, deploymentName, t]
    )

    const navItems: INavItem[] = useMemo(
        () =>
            [
                {
                    title: t('overview'),
                    path: `/clusters/${clusterName}/deployments/${deploymentName}`,
                    icon: RiSurveyLine,
                },
                {
                    title: t('replicas'),
                    path: `/clusters/${clusterName}/deployments/${deploymentName}/replicas`,
                    icon: VscServerProcess,
                },
                {
                    title: t('view log'),
                    path: `/clusters/${clusterName}/deployments/${deploymentName}/log`,
                    icon: FaJournalWhills,
                    disabled: !hasLogging,
                    helpMessage: !hasLogging ? t('please install yatai component first', [t('logging')]) : undefined,
                },
                {
                    title: t('monitor'),
                    path: `/clusters/${clusterName}/deployments/${deploymentName}/monitor`,
                    icon: AiOutlineDashboard,
                    disabled: !hasMonitoring,
                    helpMessage: !hasMonitoring
                        ? t('please install yatai component first', [t('monitoring')])
                        : undefined,
                },
                {
                    title: t('sth list', [t('revision')]),
                    path: `/clusters/${clusterName}/deployments/${deploymentName}/revisions`,
                    icon: resourceIconMapping.deployment_revision,
                },
            ] as INavItem[],
        [clusterName, deploymentName, hasLogging, hasMonitoring, t]
    )

    const [, theme] = useStyletron()
    const [isTerminateDeploymentModalOpen, setIsTerminateDeploymentModalOpen] = useState(false)
    const [isDeleteDeploymentModalOpen, setIsDeleteDeploymentModalOpen] = useState(false)

    const history = useHistory()

    const isTerminated = deployment?.status === 'terminated' || deployment?.status === 'terminating'

    return (
        <BaseSubLayout
            header={
                <Card>
                    <div
                        style={{
                            display: 'flex',
                            alignItems: 'center',
                        }}
                    >
                        <div
                            style={{
                                flexShrink: 0,
                                display: 'flex',
                                alignItems: 'center',
                                fontSize: '18px',
                                gap: 10,
                            }}
                        >
                            {React.createElement(resourceIconMapping.deployment, { size: 14 })}
                            <div>{deploymentName}</div>
                        </div>
                        <div
                            style={{
                                flexGrow: 1,
                            }}
                        />
                        <div
                            style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: 20,
                            }}
                        >
                            <DeploymentStatusTag status={deployment?.status ?? 'unknown'} />
                            {!isTerminated && (
                                <Button
                                    size='compact'
                                    overrides={{
                                        Root: {
                                            style: {
                                                ':hover': {
                                                    background: theme.colors.negative,
                                                    color: theme.colors.white,
                                                },
                                            },
                                        },
                                    }}
                                    onClick={() => setIsTerminateDeploymentModalOpen(true)}
                                >
                                    {t('terminate')}
                                </Button>
                            )}
                            {isTerminated && (
                                <Button
                                    size='compact'
                                    overrides={{
                                        Root: {
                                            style: {
                                                ':hover': {
                                                    background: theme.colors.negative,
                                                    color: theme.colors.white,
                                                },
                                            },
                                        },
                                    }}
                                    onClick={() => setIsDeleteDeploymentModalOpen(true)}
                                >
                                    {t('delete')}
                                </Button>
                            )}
                            <Button onClick={() => setIsCreateDeploymentRevisionOpen(true)} size='compact'>
                                {isTerminated ? t('restore') : t('update')}
                            </Button>
                        </div>
                    </div>
                    <Modal
                        isOpen={isCreateDeploymentRevisionOpen}
                        onClose={() => setIsCreateDeploymentRevisionOpen(false)}
                        closeable
                        animate
                        autoFocus
                    >
                        <ModalHeader>{t('update sth', [t('deployment')])}</ModalHeader>
                        <ModalBody>
                            <DeploymentForm
                                clusterName={clusterName}
                                deployment={deployment}
                                deploymentRevision={deployment?.latest_revision}
                                onSubmit={handleCreateDeploymentRevision}
                            />
                        </ModalBody>
                    </Modal>
                    <Modal
                        isOpen={isTerminateDeploymentModalOpen}
                        onClose={() => setIsTerminateDeploymentModalOpen(false)}
                        closeable
                        animate
                        autoFocus
                    >
                        <ModalHeader>
                            <div
                                style={{
                                    color: theme.colors.negative,
                                }}
                            >
                                {t('terminate sth', [t('deployment')])}
                            </div>
                        </ModalHeader>
                        <ModalBody>
                            <DoubleCheckForm
                                tips={
                                    <>
                                        <p>{t('terminate deployment tips')}</p>
                                        <p>
                                            <span>
                                                {t('double check to be continued tips prefix', [t('deployment')])}
                                            </span>
                                            <code
                                                style={{
                                                    padding: '2px 3px',
                                                    border: `1px solid ${color(theme.colors.warning400)
                                                        .lighten(0.3)
                                                        .string()}`,
                                                    background: color(theme.colors.warning100).lighten(0.1).string(),
                                                    borderRadius: '3px',
                                                    fontSize: '12px',
                                                }}
                                            >
                                                {deploymentName}
                                            </code>
                                            <span>{t('double check to be continued tips suffix')}</span>
                                        </p>
                                    </>
                                }
                                expected={deploymentName}
                                buttonLabel={t('terminate')}
                                onSubmit={async () => {
                                    await terminateDeployment(clusterName, deploymentName)
                                    setIsTerminateDeploymentModalOpen(false)
                                    deploymentInfo.refetch()
                                }}
                            />
                        </ModalBody>
                    </Modal>
                    <Modal
                        isOpen={isDeleteDeploymentModalOpen}
                        onClose={() => setIsDeleteDeploymentModalOpen(false)}
                        closeable
                        animate
                        autoFocus
                    >
                        <ModalHeader>
                            <div
                                style={{
                                    color: theme.colors.negative,
                                }}
                            >
                                {t('delete sth', [t('deployment')])}
                            </div>
                        </ModalHeader>
                        <ModalBody>
                            <DoubleCheckForm
                                tips={
                                    <>
                                        <p>
                                            <span>{t('delete deployment tips prefix')}</span>
                                            <b>{t('delete deployment tips highlight')}</b>
                                            <span>{t('delete deployment tips suffix')}</span>
                                        </p>
                                        <p>
                                            <span>
                                                {t('double check to be continued tips prefix', [t('deployment')])}
                                            </span>
                                            <code
                                                style={{
                                                    padding: '2px 3px',
                                                    border: `1px solid ${color(theme.colors.warning400)
                                                        .lighten(0.3)
                                                        .string()}`,
                                                    background: color(theme.colors.warning100).lighten(0.1).string(),
                                                    borderRadius: '3px',
                                                    fontSize: '12px',
                                                }}
                                            >
                                                {deploymentName}
                                            </code>
                                            <span>{t('double check to be continued tips suffix')}</span>
                                        </p>
                                    </>
                                }
                                expected={deploymentName}
                                buttonLabel={t('delete')}
                                onSubmit={async () => {
                                    await deleteDeployment(clusterName, deploymentName)
                                    setIsDeleteDeploymentModalOpen(false)
                                    history.push('/deployments')
                                }}
                            />
                        </ModalBody>
                    </Modal>
                </Card>
            }
            breadcrumbItems={breadcrumbItems}
            navItems={navItems}
        >
            {children}
        </BaseSubLayout>
    )
}
