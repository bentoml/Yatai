import { resourceIconMapping } from '@/consts'
import { useCluster } from '@/hooks/useCluster'
import { useCurrentUser } from '@/hooks/useCurrentUser'
import { useDeployment, useDeploymentLoading } from '@/hooks/useDeployment'
import { useFetchDeployment } from '@/hooks/useFetchDeployment'
import { useFetchDeploymentRevisions } from '@/hooks/useFetchDeploymentRevisions'
import { useFetchOrganizationMembers } from '@/hooks/useFetchOrganizationMembers'
import { useFetchYataiComponents } from '@/hooks/useFetchYataiComponents'
import { useOrganization } from '@/hooks/useOrganization'
import { usePage } from '@/hooks/usePage'
import { useSubscription } from '@/hooks/useSubscription'
import useTranslation from '@/hooks/useTranslation'
import { IDeploymentFullSchema, IDeploymentSchema, IUpdateDeploymentSchema } from '@/schemas/deployment'
import { deleteDeployment, terminateDeployment, updateDeployment } from '@/services/deployment'
import { useStyletron } from 'baseui'
import { Block } from 'baseui/block'
import { Modal, ModalBody, ModalHeader } from 'baseui/modal'
import color from 'color'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { AiOutlineDashboard } from 'react-icons/ai'
import { FaJournalWhills } from 'react-icons/fa'
import { RiSurveyLine } from 'react-icons/ri'
import { VscServerProcess } from 'react-icons/vsc'
import { useQueryClient } from 'react-query'
import { useHistory, useParams } from 'react-router-dom'
import { INavItem } from './BaseSidebar'
import BaseSubLayout from './BaseSubLayout'
import Card from './Card'
import DeploymentForm from './DeploymentForm'
import DeploymentStatusTag from './DeploymentStatusTag'
import DoubleCheckForm from './DoubleCheckForm'
import TooltipButton from './TooltipButton'

export interface IDeploymentLayoutProps {
    children: React.ReactNode
}

export default function DeploymentLayout({ children }: IDeploymentLayoutProps) {
    const { clusterName, kubeNamespace, deploymentName } =
        useParams<{ clusterName: string; kubeNamespace: string; deploymentName: string }>()
    const { queryKey, deploymentInfo } = useFetchDeployment(clusterName, kubeNamespace, deploymentName)
    const { deployment, setDeployment } = useDeployment()
    const { organization, setOrganization } = useOrganization()
    const { cluster, setCluster } = useCluster()
    const { setDeploymentLoading } = useDeploymentLoading()
    useEffect(() => {
        setDeploymentLoading(deploymentInfo.isLoading)
        if (deploymentInfo.isSuccess) {
            setDeployment(deploymentInfo.data)
            if (deploymentInfo.data.cluster?.uid !== cluster?.uid) {
                setCluster(deploymentInfo.data.cluster)
            }
            if (deploymentInfo.data.cluster?.organization?.uid !== organization?.uid) {
                setOrganization(deploymentInfo.data.cluster?.organization)
            }
        } else if (deploymentInfo.isLoading) {
            setDeployment(undefined)
        }
        return () => {
            setDeployment(undefined)
        }
    }, [
        cluster?.uid,
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
    const { deploymentRevisionsInfo } = useFetchDeploymentRevisions(clusterName, kubeNamespace, deploymentName, page)
    const [isCreateDeploymentRevisionOpen, setIsCreateDeploymentRevisionOpen] = useState(false)
    const handleCreateDeploymentRevision = useCallback(
        async (data: IUpdateDeploymentSchema) => {
            await updateDeployment(clusterName, kubeNamespace, deploymentName, data)
            await deploymentInfo.refetch()
            await deploymentRevisionsInfo.refetch()
            setIsCreateDeploymentRevisionOpen(false)
        },
        [clusterName, kubeNamespace, deploymentName, deploymentInfo, deploymentRevisionsInfo]
    )

    const breadcrumbItems: INavItem[] = useMemo(
        () => [
            {
                title: t('deployments'),
                path: '/deployments',
                icon: resourceIconMapping.deployment,
            },
            {
                title: deploymentName,
                path: `/clusters/${clusterName}/namespaces/${kubeNamespace}/deployments/${deploymentName}`,
            },
        ],
        [clusterName, deploymentName, kubeNamespace, t]
    )

    const navItems: INavItem[] = useMemo(
        () =>
            [
                {
                    title: t('overview'),
                    path: `/clusters/${clusterName}/namespaces/${kubeNamespace}/deployments/${deploymentName}`,
                    icon: RiSurveyLine,
                },
                {
                    title: t('replicas'),
                    path: `/clusters/${clusterName}/namespaces/${kubeNamespace}/deployments/${deploymentName}/replicas`,
                    icon: VscServerProcess,
                },
                {
                    title: t('view log'),
                    path: `/clusters/${clusterName}/namespaces/${kubeNamespace}/deployments/${deploymentName}/log`,
                    icon: FaJournalWhills,
                    disabled: !hasLogging,
                    helpMessage: !hasLogging ? t('please install yatai component first', [t('logging')]) : undefined,
                },
                {
                    title: t('monitor'),
                    path: `/clusters/${clusterName}/namespaces/${kubeNamespace}/deployments/${deploymentName}/monitor`,
                    icon: AiOutlineDashboard,
                    disabled: !hasMonitoring,
                    helpMessage: !hasMonitoring
                        ? t('please install yatai component first', [t('monitoring')])
                        : undefined,
                },
                {
                    title: t('revisions'),
                    path: `/clusters/${clusterName}/namespaces/${kubeNamespace}/deployments/${deploymentName}/revisions`,
                    icon: resourceIconMapping.deployment_revision,
                },
            ] as INavItem[],
        [clusterName, deploymentName, hasLogging, hasMonitoring, kubeNamespace, t]
    )

    const [, theme] = useStyletron()
    const [isTerminateDeploymentModalOpen, setIsTerminateDeploymentModalOpen] = useState(false)
    const [isDeleteDeploymentModalOpen, setIsDeleteDeploymentModalOpen] = useState(false)

    const history = useHistory()

    const isTerminated = deployment?.status === 'terminated' || deployment?.status === 'terminating'

    const membersInfo = useFetchOrganizationMembers()
    const { currentUser } = useCurrentUser()
    const hasOperationPermission = useMemo(() => {
        return membersInfo.data?.find((x) => x.user.uid === currentUser?.uid && x.role === 'admin') !== undefined
    }, [currentUser?.uid, membersInfo.data])

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
                                <TooltipButton
                                    tooltip={
                                        !hasOperationPermission
                                            ? () => {
                                                  return (
                                                      <Block width={['200px', '400px']}>
                                                          {t(
                                                              'Only the administrator has permission to operate deployments, please contact the administrator'
                                                          )}
                                                      </Block>
                                                  )
                                              }
                                            : undefined
                                    }
                                    disabled={!hasOperationPermission}
                                    size='compact'
                                    overrides={
                                        hasOperationPermission
                                            ? {
                                                  Root: {
                                                      style: {
                                                          ':hover': {
                                                              background: theme.colors.negative,
                                                              color: theme.colors.white,
                                                          },
                                                      },
                                                  },
                                              }
                                            : undefined
                                    }
                                    onClick={() => setIsTerminateDeploymentModalOpen(true)}
                                >
                                    {t('terminate')}
                                </TooltipButton>
                            )}
                            {isTerminated && (
                                <TooltipButton
                                    tooltip={
                                        !hasOperationPermission
                                            ? () => {
                                                  return (
                                                      <Block width={['200px', '400px']}>
                                                          {t(
                                                              'Only the administrator has permission to operate deployments, please contact the administrator'
                                                          )}
                                                      </Block>
                                                  )
                                              }
                                            : undefined
                                    }
                                    disabled={!hasOperationPermission}
                                    size='compact'
                                    overrides={
                                        hasOperationPermission
                                            ? {
                                                  Root: {
                                                      style: {
                                                          ':hover': {
                                                              background: theme.colors.negative,
                                                              color: theme.colors.white,
                                                          },
                                                      },
                                                  },
                                              }
                                            : undefined
                                    }
                                    onClick={() => setIsDeleteDeploymentModalOpen(true)}
                                >
                                    {t('delete')}
                                </TooltipButton>
                            )}
                            <TooltipButton
                                tooltip={
                                    !hasOperationPermission
                                        ? () => {
                                              return (
                                                  <Block width={['100px', '200px', '400px']}>
                                                      {t(
                                                          'Only the administrator has permission to operate deployments, please contact the administrator'
                                                      )}
                                                  </Block>
                                              )
                                          }
                                        : undefined
                                }
                                disabled={!hasOperationPermission}
                                onClick={() =>
                                    history.push(
                                        `/clusters/${clusterName}/namespaces/${kubeNamespace}/deployments/${deploymentName}/edit`
                                    )
                                }
                                size='compact'
                            >
                                {isTerminated ? t('restore') : t('update')}
                            </TooltipButton>
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
                                    await terminateDeployment(clusterName, kubeNamespace, deploymentName)
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
                                    await deleteDeployment(clusterName, kubeNamespace, deploymentName)
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
