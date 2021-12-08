import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useQuery, useQueryClient } from 'react-query'
import Card from '@/components/Card'
import { createBento, listBentos } from '@/services/bento'
import { usePage } from '@/hooks/usePage'
import { IBentoSchema, ICreateBentoSchema } from '@/schemas/bento'
import BentoForm from '@/components/BentoForm'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import User from '@/components/User'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import { StatefulTooltip, PLACEMENT } from 'baseui/tooltip'
import Table from '@/components/Table'
import { Link } from 'react-router-dom'
import { resourceIconMapping } from '@/consts'
import { useSubscription } from '@/hooks/useSubscription'
import { IListSchema } from '@/schemas/list'
import BentoImageBuildStatusTag from '@/components/BentoImageBuildStatus'
import qs from 'qs'
import ReactTimeAgo from 'react-time-ago'

export interface IBentoListCardProps {
    bentoRepositoryName: string
}

export default function BentoListCard({ bentoRepositoryName }: IBentoListCardProps) {
    const [page] = usePage()
    const queryKey = `fetchBentos:${bentoRepositoryName}:${qs.stringify(page)}`
    const bentosInfo = useQuery(queryKey, () => listBentos(bentoRepositoryName, page))
    const [isCreateBentoVersionOpen, setIsCreateBentoVersionOpen] = useState(false)
    const handleCreateBentoVersion = useCallback(
        async (data: ICreateBentoSchema) => {
            await createBento(bentoRepositoryName, data)
            await bentosInfo.refetch()
            setIsCreateBentoVersionOpen(false)
        },
        [bentoRepositoryName, bentosInfo]
    )
    const [t] = useTranslation()

    const uids = useMemo(() => bentosInfo.data?.items.map((bento) => bento.uid) ?? [], [bentosInfo.data?.items])
    const queryClient = useQueryClient()
    const subscribeCb = useCallback(
        (bentoVersion: IBentoSchema) => {
            queryClient.setQueryData(queryKey, (oldData?: IListSchema<IBentoSchema>): IListSchema<IBentoSchema> => {
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
            })
        },
        [queryClient, queryKey]
    )
    const { subscribe, unsubscribe } = useSubscription()

    useEffect(() => {
        subscribe({
            resourceType: 'bento',
            resourceUids: uids,
            cb: subscribeCb,
        })
        return () => {
            unsubscribe({
                resourceType: 'bento',
                resourceUids: uids,
                cb: subscribeCb,
            })
        }
    }, [subscribe, subscribeCb, uids, unsubscribe])

    return (
        <Card
            title={t('sth list', [t('bento')])}
            titleIcon={resourceIconMapping.bento}
            extra={
                <Button size={ButtonSize.compact} onClick={() => setIsCreateBentoVersionOpen(true)}>
                    {t('create')}
                </Button>
            }
        >
            <Table
                isLoading={bentosInfo.isLoading}
                columns={[t('name'), t('image build status'), t('description'), t('creator'), t('Time since Creation')]}
                data={
                    bentosInfo.data?.items.map((bento) => [
                        <Link key={bento.uid} to={`/bento_repositories/${bentoRepositoryName}/bentos/${bento.version}`}>
                            {bento.version}
                        </Link>,
                        <BentoImageBuildStatusTag key={bento.uid} status={bento.image_build_status} />,
                        bento.description,
                        bento.creator && <User user={bento.creator} />,
                        bento?.created_at && (
                            <StatefulTooltip placement={PLACEMENT.bottom} content={() => bento?.created_at}>
                                <ReactTimeAgo date={new Date(bento.created_at)} timeStyle='round' locale='en-US' />
                            </StatefulTooltip>
                        ),
                    ]) ?? []
                }
                paginationProps={{
                    start: bentosInfo.data?.start,
                    count: bentosInfo.data?.count,
                    total: bentosInfo.data?.total,
                    afterPageChange: () => {
                        bentosInfo.refetch()
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
                    <BentoForm onSubmit={handleCreateBentoVersion} />
                </ModalBody>
            </Modal>
        </Card>
    )
}
