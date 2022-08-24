import { useCallback, useEffect, useMemo, useState } from 'react'
import { useQuery, useQueryClient } from 'react-query'
import Card from '@/components/Card'
import { createDeployment, listClusterDeployments, listOrganizationDeployments } from '@/services/deployment'
import { usePage } from '@/hooks/usePage'
import { ICreateDeploymentSchema, IDeploymentSchema } from '@/schemas/deployment'
import DeploymentForm from '@/components/DeploymentForm'
import useTranslation from '@/hooks/useTranslation'
import User from '@/components/User'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import Table from '@/components/Table'
import { useHistory } from 'react-router-dom'
import { resourceIconMapping } from '@/consts'
import { IListSchema } from '@/schemas/list'
import { useSubscription } from '@/hooks/useSubscription'
import DeploymentStatusTag from '@/components/DeploymentStatusTag'
import { useFetchOrganizationMembers } from '@/hooks/useFetchOrganizationMembers'
import qs from 'qs'
import { useQ } from '@/hooks/useQ'
import FilterBar from '@/components/FilterBar'
import { useFetchClusters } from '@/hooks/useFetchClusters'
import { useCurrentUser } from '@/hooks/useCurrentUser'
import { useOrganization } from '@/hooks/useOrganization'
import FilterInput from './FilterInput'
import Time from './Time'
import Link from './Link'
import TooltipButton from './TooltipButton'

export interface IDeploymentListCardProps {
    clusterName?: string
}

export default function DeploymentListCard({ clusterName }: IDeploymentListCardProps) {
    const { q, updateQ } = useQ()
    const membersInfo = useFetchOrganizationMembers()
    const { currentUser } = useCurrentUser()
    const clustersInfo = useFetchClusters({
        start: 0,
        count: 1000,
    })
    const [page] = usePage()
    const { organization } = useOrganization()
    const queryKey = `fetchClusterDeployments:${organization?.name}:${clusterName ?? ''}:${qs.stringify(page)}`
    const deploymentsInfo = useQuery(queryKey, () =>
        clusterName ? listClusterDeployments(clusterName, page) : listOrganizationDeployments(page)
    )
    const [isCreateDeploymentOpen, setIsCreateDeploymentOpen] = useState(false)
    const handleCreateDeployment = useCallback(
        async (data: ICreateDeploymentSchema) => {
            if (!data.cluster_name) {
                return
            }
            await createDeployment(data.cluster_name, data)
            await deploymentsInfo.refetch()
            setIsCreateDeploymentOpen(false)
        },
        [deploymentsInfo]
    )
    const [t] = useTranslation()

    const uids = useMemo(
        () => deploymentsInfo.data?.items.map((deployment) => deployment.uid) ?? [],
        [deploymentsInfo.data?.items]
    )
    const queryClient = useQueryClient()
    const subscribeCb = useCallback(
        (deployment: IDeploymentSchema) => {
            queryClient.setQueryData(
                queryKey,
                (oldData?: IListSchema<IDeploymentSchema>): IListSchema<IDeploymentSchema> => {
                    if (!oldData) {
                        return {
                            start: 0,
                            count: 0,
                            total: 0,
                            items: [],
                        }
                    }
                    return {
                        ...oldData,
                        items: oldData.items.map((oldDeployment) => {
                            if (oldDeployment.uid === deployment.uid) {
                                return {
                                    ...oldDeployment,
                                    ...deployment,
                                }
                            }
                            return oldDeployment
                        }),
                    }
                }
            )
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

    const history = useHistory()

    const hasOperationPermission = useMemo(
        () =>
            membersInfo?.data?.find((m) => m.user.uid === currentUser?.uid && m.role === 'admin') !== undefined ||
            currentUser?.uid === organization?.creator?.uid ||
            currentUser?.is_super_admin,
        [currentUser?.is_super_admin, currentUser?.uid, membersInfo?.data, organization?.creator?.uid]
    )

    return (
        <Card
            title={t('deployments')}
            titleIcon={resourceIconMapping.deployment}
            middle={
                <div
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        flexGrow: 1,
                    }}
                >
                    <div
                        style={{
                            width: 100,
                            flexGrow: 1,
                        }}
                    />
                    <div
                        style={{
                            flexGrow: 2,
                            flexShrink: 0,
                            maxWidth: 1200,
                        }}
                    >
                        <FilterInput
                            filterConditions={[
                                {
                                    qStr: 'creator:@me',
                                    label: t('the deployments I created'),
                                },
                                {
                                    qStr: 'last_updater:@me',
                                    label: t('my last updated deployments'),
                                },
                            ]}
                        />
                    </div>
                </div>
            }
            extra={
                <TooltipButton
                    tooltip={
                        !hasOperationPermission
                            ? () =>
                                  t(
                                      'Only the administrator has permission to create deployments, please contact the administrator'
                                  )
                            : undefined
                    }
                    disabled={!hasOperationPermission}
                    size='compact'
                    onClick={() => history.push('/new_deployment')}
                >
                    {t('create')}
                </TooltipButton>
            }
        >
            <FilterBar
                filters={[
                    {
                        showInput: true,
                        multiple: true,
                        options:
                            clustersInfo.data?.items.map((cluster) => ({
                                id: cluster.name,
                                label: cluster.name,
                            })) ?? [],

                        value: ((q.cluster as string[] | undefined) ?? []).map((v) => ({
                            id: v,
                        })),
                        onChange: ({ value }) => {
                            updateQ({
                                cluster: value.map((v) => String(v.id ?? '')),
                            })
                        },
                        label: t('cluster'),
                        description: t('filter by sth', [t('cluster')]),
                    },
                    {
                        showInput: true,
                        multiple: true,
                        options:
                            membersInfo.data?.map(({ user }) => ({
                                id: user.name,
                                label: <User user={user} />,
                            })) ?? [],

                        value: ((q.last_updater as string[] | undefined) ?? []).map((v) => ({
                            id: v,
                        })),
                        onChange: ({ value }) => {
                            updateQ({
                                last_updater: value.map((v) => String(v.id ?? '')),
                            })
                        },
                        label: t('Created At'),
                        description: t('filter by sth', [t('Created At')]),
                    },
                    {
                        options: [
                            {
                                id: 'updated_at-desc',
                                label: t('newest update'),
                            },
                            {
                                id: 'updated_at-asc',
                                label: t('oldest update'),
                            },
                        ],
                        value: ((q.sort as string[] | undefined) ?? []).map((v) => ({
                            id: v,
                        })),
                        onChange: ({ value }) => {
                            updateQ({
                                sort: value.map((v) => String(v.id ?? '')),
                            })
                        },
                        label: t('sort'),
                    },
                ]}
            />
            <Table
                isLoading={deploymentsInfo.isLoading}
                columns={[
                    t('name'),
                    clusterName ? undefined : t('cluster'),
                    t('status'),
                    t('creator'),
                    t('Time since Creation'),
                    t('updated_at'),
                ]}
                data={
                    deploymentsInfo.data?.items.map((deployment) => {
                        return [
                            <Link
                                key={deployment.uid}
                                href={`/clusters/${deployment.cluster?.name}/namespaces/${deployment.kube_namespace}/deployments/${deployment.name}`}
                            >
                                {deployment.name}
                            </Link>,
                            clusterName ? undefined : (
                                <Link key={deployment.cluster?.uid} href={`/clusters/${deployment.cluster?.name}`}>
                                    {deployment.cluster?.name}
                                </Link>
                            ),
                            <DeploymentStatusTag key={deployment.uid} status={deployment.status} />,
                            deployment?.creator && <User user={deployment.creator} />,
                            deployment?.created_at && <Time time={deployment.created_at} />,
                            deployment?.latest_revision && <Time time={deployment.latest_revision.updated_at} />,
                        ]
                    }) ?? []
                }
                paginationProps={{
                    start: deploymentsInfo.data?.start,
                    count: deploymentsInfo.data?.count,
                    total: deploymentsInfo.data?.total,
                    afterPageChange: () => {
                        deploymentsInfo.refetch()
                    },
                }}
            />
            <Modal
                isOpen={isCreateDeploymentOpen}
                onClose={() => setIsCreateDeploymentOpen(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('create sth', [t('deployment')])}</ModalHeader>
                <ModalBody>
                    <DeploymentForm onSubmit={handleCreateDeployment} clusterName={clusterName} />
                </ModalBody>
            </Modal>
        </Card>
    )
}
