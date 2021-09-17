import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useQuery, useQueryClient } from 'react-query'
import Card from '@/components/Card'
import { createDeployment, listClusterDeployments, listOrganizationDeployments } from '@/services/deployment'
import { usePage } from '@/hooks/usePage'
import { DeploymentStatus, ICreateDeploymentSchema, IDeploymentSchema } from '@/schemas/deployment'
import DeploymentForm from '@/components/DeploymentForm'
import { formatTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import User from '@/components/User'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import Table from '@/components/Table'
import { Link } from 'react-router-dom'
import { resourceIconMapping } from '@/consts'
import { Tag, KIND as TagKind, VARIANT as TagVariant } from 'baseui/tag'
import { IListSchema } from '@/schemas/list'
import { useSubscription } from '@/hooks/useSubscription'
import { StyledSpinnerNext } from 'baseui/spinner'

export interface IDeploymentListCardProps {
    orgName: string
    clusterName?: string
}

export default function DeploymentListCard({ orgName, clusterName }: IDeploymentListCardProps) {
    const [page, setPage] = usePage()
    const queryKey = `fetchClusterDeployments:${orgName}:${clusterName ?? ''}`
    const deploymentsInfo = useQuery(queryKey, () =>
        clusterName ? listClusterDeployments(orgName, clusterName, page) : listOrganizationDeployments(orgName, page)
    )
    const [isCreateDeploymentOpen, setIsCreateDeploymentOpen] = useState(false)
    const handleCreateDeployment = useCallback(
        async (data: ICreateDeploymentSchema) => {
            if (!data.cluster_name) {
                return
            }
            await createDeployment(orgName, data.cluster_name, data)
            await deploymentsInfo.refetch()
            setIsCreateDeploymentOpen(false)
        },
        [deploymentsInfo, orgName]
    )
    const [t] = useTranslation()

    const statusColorMap: Record<DeploymentStatus, keyof TagKind> = useMemo(() => {
        return {
            'unknown': TagKind.black,
            'non-deployed': TagKind.primary,
            'running': TagKind.positive,
            'unhealthy': TagKind.warning,
            'failed': TagKind.negative,
            'deploying': TagKind.accent,
        }
    }, [])

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
            extra={
                <Button size={ButtonSize.compact} onClick={() => setIsCreateDeploymentOpen(true)}>
                    {t('create')}
                </Button>
            }
        >
            <Table
                isLoading={deploymentsInfo.isLoading}
                columns={[
                    t('name'),
                    clusterName ? undefined : t('cluster'),
                    t('status'),
                    t('creator'),
                    t('created_at'),
                ]}
                data={
                    deploymentsInfo.data?.items.map((deployment) => [
                        <Link
                            key={deployment.uid}
                            to={`/orgs/${orgName}/clusters/${deployment.cluster?.name}/deployments/${deployment.name}`}
                        >
                            {deployment.name}
                        </Link>,
                        clusterName ? undefined : (
                            <Link
                                key={deployment.cluster?.uid}
                                to={`/orgs/${orgName}/clusters/${deployment.cluster?.name}`}
                            >
                                {deployment.cluster?.name}
                            </Link>
                        ),
                        <Tag
                            key={deployment.uid}
                            closeable={false}
                            variant={TagVariant.light}
                            kind={statusColorMap[deployment.status]}
                        >
                            <div
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: 4,
                                }}
                            >
                                {['deploying'].indexOf(deployment.status) >= 0 && <StyledSpinnerNext $size={100} />}
                                {t(deployment.status)}
                            </div>
                        </Tag>,
                        deployment.creator && <User user={deployment.creator} />,
                        formatTime(deployment.created_at),
                    ]) ?? []
                }
                paginationProps={{
                    start: deploymentsInfo.data?.start,
                    count: deploymentsInfo.data?.count,
                    total: deploymentsInfo.data?.total,
                    onPageChange: ({ nextPage }) => {
                        setPage({
                            ...page,
                            start: nextPage * page.count,
                        })
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
                    <DeploymentForm onSubmit={handleCreateDeployment} orgName={orgName} clusterName={clusterName} />
                </ModalBody>
            </Modal>
        </Card>
    )
}
