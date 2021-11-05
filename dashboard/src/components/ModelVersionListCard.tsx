import { resourceIconMapping } from '@/consts'
import { usePage } from '@/hooks/usePage'
import { useSubscription } from '@/hooks/useSubscription'
import useTranslation from '@/hooks/useTranslation'
import { IListSchema } from '@/schemas/list'
import { ICreateModelVersionSchema, IModelVersionSchema } from '@/schemas/model_version'
import { createModelVersion, listModelVersions } from '@/services/model_version'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import React, { useCallback, useMemo, useState } from 'react'
import { useQuery, useQueryClient } from 'react-query'
import Card from '@/components/Card'
import Table from '@/components/Table'
import User from '@/components/User'
import { formatTime } from '@/utils/datetime'
import { Link } from 'react-router-dom'
import { Modal, ModalBody, ModalHeader } from 'baseui/modal'

export interface IModelVersionListCardProps {
    modelName: string
}

export default function ModelVersionListCard({ modelName }: IModelVersionListCardProps) {
    const [page, setPage] = usePage()
    const queryKey = `fetchClusterModelVersions:${modelName}`
    const modelVersionsInfo = useQuery(queryKey, () => listModelVersions(modelName, page))
    const [isCreateModelVersionOpen, setIsCreateModelVersionOpen] = useState(false)
    // eslint-disable-next-line
    const handleCreateModelVersionOpen = useCallback(
        async (data: ICreateModelVersionSchema) => {
            await createModelVersion(modelName, data)
            await modelVersionsInfo.refetch()
            setIsCreateModelVersionOpen(false)
        },
        [modelName, modelVersionsInfo]
    )
    const [t] = useTranslation()

    // eslint-disable-next-line
    const uids = useMemo(
        () => modelVersionsInfo.data?.items.map((modelVersion) => modelVersion.uid ?? []),
        [modelVersionsInfo.data?.items]
    )
    const queryClient = useQueryClient()
    // eslint-disable-next-line
    const subscribeCb = useCallback(
        (modelVersion: IModelVersionSchema) => {
            queryClient.setQueryData(
                queryKey,
                (prevData?: IListSchema<IModelVersionSchema>): IListSchema<IModelVersionSchema> => {
                    if (!prevData) {
                        return {
                            start: 0,
                            count: 0,
                            total: 0,
                            items: [],
                        }
                    }
                    return {
                        ...prevData,
                        items: prevData.items.map((item) => {
                            if (item.uid === modelVersion.uid) {
                                return {
                                    ...item,
                                    ...modelVersion,
                                }
                            }
                            return item
                        }),
                    }
                }
            )
        },
        [queryClient, queryKey]
    )
    // eslint-disable-next-line
    const { subscribe, unsubscribe } = useSubscription()

    // useEffect(() => {
    //     subscribe({
    //         resourceType: 'model_version',
    //         resourceUids: uids,
    //         cb: subscribeCb,
    //     })
    //     return () => {
    //         unsubscribe({
    //             resourceType: 'model_version',
    //             resourceUids: uids,
    //             cb: subscribeCb,
    //         })
    //     }
    // }, [subscribe, subscribeCb, uids, unsubscribe])

    return (
        <Card
            title={t('sth list', [t('version')])}
            titleIcon={resourceIconMapping.model_version}
            extra={
                <Button size={ButtonSize.compact} onClick={() => setIsCreateModelVersionOpen(true)}>
                    {t('create')}
                </Button>
            }
        >
            <Table
                isLoading={modelVersionsInfo.isLoading}
                columns={[t('version'), t('creator'), t('created_at')]}
                data={
                    modelVersionsInfo.data?.items.map((modelVersion) => [
                        <Link key={modelVersion.uid} to={`/models/${modelName}/versions/${modelVersion.version}`}>
                            {modelVersion.version}
                        </Link>,
                        modelVersion.creator && <User user={modelVersion.creator} />,
                        formatTime(modelVersion.created_at),
                    ]) ?? []
                }
                paginationProps={{
                    start: modelVersionsInfo.data?.start,
                    count: modelVersionsInfo.data?.count,
                    total: modelVersionsInfo.data?.total,
                    onPageChange: ({ nextPage }) => {
                        setPage({
                            ...page,
                            start: nextPage * page.count,
                        })
                        modelVersionsInfo.refetch()
                    },
                }}
            />
            <Modal
                isOpen={isCreateModelVersionOpen}
                onClose={() => setIsCreateModelVersionOpen(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('create sth', [t('version')])}</ModalHeader>
                <ModalBody>Modal form body</ModalBody>
            </Modal>
        </Card>
    )
}
