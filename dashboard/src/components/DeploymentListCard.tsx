import { useCallback, useEffect, useMemo, useState } from 'react'
import { useQuery, useQueryClient } from 'react-query'
import Card from '@/components/Card'
import { createDeployment, listClusterDeployments, listOrganizationDeployments } from '@/services/deployment'
import { usePage } from '@/hooks/usePage'
import { ICreateDeploymentSchema, IDeploymentSchema } from '@/schemas/deployment'
import DeploymentForm from '@/components/DeploymentForm'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import User from '@/components/User'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import Table from '@/components/Table'
import { Link } from 'react-router-dom'
import { resourceIconMapping } from '@/consts'
import { IListSchema } from '@/schemas/list'
import { useSubscription } from '@/hooks/useSubscription'
import DeploymentStatusTag from '@/components/DeploymentStatusTag'
import { useFetchOrganizationMembers } from '@/hooks/useFetchOrganizationMembers'
import qs from 'qs'
import { useQ } from '@/hooks/useQ'
import FilterBar from '@/components/FilterBar'
import { useFetchClusters } from '@/hooks/useFetchClusters'
import { StatefulTooltip, PLACEMENT } from 'baseui/tooltip'
import ReactTimeAgo from 'react-time-ago'
import FilterInput from './FilterInput'

export interface IDeploymentListCardProps {
    clusterName?: string
}

export default function DeploymentListCard({ clusterName }: IDeploymentListCardProps) {
    const { q, updateQ } = useQ()
    const membersInfo = useFetchOrganizationMembers()
    const clustersInfo = useFetchClusters({
        start: 0,
        count: 1000,
    })
    const [page] = usePage()
    const queryKey = `fetchClusterDeployments:${clusterName ?? ''}:${qs.stringify(page)}`
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

    return (
        <Card
            title={t('sth list', [t('deployment')])}
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
                <Button size={ButtonSize.compact} onClick={() => setIsCreateDeploymentOpen(true)}>
                    {t('create')}
                </Button>
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
                        console.log(new Date(deployment.created_at))
                        return [
                            <Link
                                key={deployment.uid}
                                to={`/clusters/${deployment.cluster?.name}/deployments/${deployment.name}`}
                            >
                                {deployment.name}
                            </Link>,
                            clusterName ? undefined : (
                                <Link key={deployment.cluster?.uid} to={`/clusters/${deployment.cluster?.name}`}>
                                    {deployment.cluster?.name}
                                </Link>
                            ),
                            <DeploymentStatusTag key={deployment.uid} status={deployment.status} />,
                            deployment?.creator && <User user={deployment.creator} />,
                            deployment?.created_at && (
                                <StatefulTooltip placement={PLACEMENT.bottom} content={() => deployment.created_at}>
                                    <ReactTimeAgo
                                        date={new Date(deployment.created_at)}
                                        timeStyle='round'
                                        locale='en-US'
                                    />
                                </StatefulTooltip>
                            ),
                            deployment?.latest_revision && (
                                <StatefulTooltip
                                    placement={PLACEMENT.bottom}
                                    content={() => deployment.latest_revision?.created_at}
                                >
                                    <ReactTimeAgo
                                        date={new Date(deployment.latest_revision?.created_at)}
                                        locale='en-US'
                                        timeStyle='round'
                                    />
                                </StatefulTooltip>
                            ),
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
