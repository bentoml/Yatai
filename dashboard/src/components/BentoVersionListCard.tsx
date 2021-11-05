import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useQuery, useQueryClient } from 'react-query'
import Card from '@/components/Card'
import { createBentoVersion, listBentoVersions } from '@/services/bento_version'
import { usePage } from '@/hooks/usePage'
import { IBentoVersionSchema, ICreateBentoVersionSchema } from '@/schemas/bento_version'
import BentoVersionForm from '@/components/BentoVersionForm'
import { formatTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import User from '@/components/User'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import Table from '@/components/Table'
import { Link } from 'react-router-dom'
import { resourceIconMapping } from '@/consts'
import { useSubscription } from '@/hooks/useSubscription'
import { IListSchema } from '@/schemas/list'
import BentoVersionImageBuildStatusTag from '@/components/BentoVersionImageBuildStatus'

export interface IBentoVersionListCardProps {
    bentoName: string
}

export default function BentoVersionListCard({ bentoName }: IBentoVersionListCardProps) {
    const [page, setPage] = usePage()
    const queryKey = `fetchClusterBentoVersions:${bentoName}`
    const bentoVersionsInfo = useQuery(queryKey, () => listBentoVersions(bentoName, page))
    const [isCreateBentoVersionOpen, setIsCreateBentoVersionOpen] = useState(false)
    const handleCreateBentoVersion = useCallback(
        async (data: ICreateBentoVersionSchema) => {
            await createBentoVersion(bentoName, data)
            await bentoVersionsInfo.refetch()
            setIsCreateBentoVersionOpen(false)
        },
        [bentoName, bentoVersionsInfo]
    )
    const [t] = useTranslation()

    const uids = useMemo(
        () => bentoVersionsInfo.data?.items.map((bentoVersion) => bentoVersion.uid) ?? [],
        [bentoVersionsInfo.data?.items]
    )
    const queryClient = useQueryClient()
    const subscribeCb = useCallback(
        (bentoVersion: IBentoVersionSchema) => {
            queryClient.setQueryData(
                queryKey,
                (oldData?: IListSchema<IBentoVersionSchema>): IListSchema<IBentoVersionSchema> => {
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
                        items: oldData.items.map((oldBentoVersion) => {
                            if (oldBentoVersion.uid === bentoVersion.uid) {
                                return {
                                    ...oldBentoVersion,
                                    ...bentoVersion,
                                }
                            }
                            return oldBentoVersion
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
            resourceType: 'bento_version',
            resourceUids: uids,
            cb: subscribeCb,
        })
        return () => {
            unsubscribe({
                resourceType: 'bento_version',
                resourceUids: uids,
                cb: subscribeCb,
            })
        }
    }, [subscribe, subscribeCb, uids, unsubscribe])

    return (
        <Card
            title={t('sth list', [t('version')])}
            titleIcon={resourceIconMapping.bento_version}
            extra={
                <Button size={ButtonSize.compact} onClick={() => setIsCreateBentoVersionOpen(true)}>
                    {t('create')}
                </Button>
            }
        >
            <Table
                isLoading={bentoVersionsInfo.isLoading}
                columns={[t('name'), t('image build status'), t('description'), t('creator'), t('created_at')]}
                data={
                    bentoVersionsInfo.data?.items.map((bentoVersion) => [
                        <Link key={bentoVersion.uid} to={`/bentos/${bentoName}/versions/${bentoVersion.version}`}>
                            {bentoVersion.version}
                        </Link>,
                        <BentoVersionImageBuildStatusTag
                            key={bentoVersion.uid}
                            status={bentoVersion.image_build_status}
                        />,
                        bentoVersion.description,
                        bentoVersion.creator && <User user={bentoVersion.creator} />,
                        formatTime(bentoVersion.created_at),
                    ]) ?? []
                }
                paginationProps={{
                    start: bentoVersionsInfo.data?.start,
                    count: bentoVersionsInfo.data?.count,
                    total: bentoVersionsInfo.data?.total,
                    onPageChange: ({ nextPage }) => {
                        setPage({
                            ...page,
                            start: nextPage * page.count,
                        })
                        bentoVersionsInfo.refetch()
                    },
                }}
            />
            <Modal
                isOpen={isCreateBentoVersionOpen}
                onClose={() => setIsCreateBentoVersionOpen(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('create sth', [t('version')])}</ModalHeader>
                <ModalBody>
                    <BentoVersionForm onSubmit={handleCreateBentoVersion} />
                </ModalBody>
            </Modal>
        </Card>
    )
}
