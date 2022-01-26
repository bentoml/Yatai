import React, { useCallback, useState } from 'react'
import Card from '@/components/Card'
import { createCluster } from '@/services/cluster'
import { usePage } from '@/hooks/usePage'
import { ICreateClusterSchema } from '@/schemas/cluster'
import ClusterForm from '@/components/ClusterForm'
import { formatDateTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import User from '@/components/User'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import Table from '@/components/Table'
import { useFetchClusters } from '@/hooks/useFetchClusters'
import { resourceIconMapping } from '@/consts'
import Link from './Link'

export default function ClusterListCard() {
    const [page] = usePage()
    const clustersInfo = useFetchClusters(page)
    const [isCreateClusterOpen, setIsCreateClusterOpen] = useState(false)
    const handleCreateCluster = useCallback(
        async (data: ICreateClusterSchema) => {
            await createCluster(data)
            await clustersInfo.refetch()
            setIsCreateClusterOpen(false)
        },
        [clustersInfo]
    )
    const [t] = useTranslation()

    return (
        <Card
            title={t('clusters')}
            titleIcon={resourceIconMapping.cluster}
            extra={
                <Button size={ButtonSize.compact} onClick={() => setIsCreateClusterOpen(true)}>
                    {t('create')}
                </Button>
            }
        >
            <Table
                isLoading={clustersInfo.isLoading}
                columns={[t('name'), t('description'), t('creator'), t('created_at')]}
                data={
                    clustersInfo.data?.items.map((cluster) => [
                        <Link key={cluster.uid} href={`/clusters/${cluster.name}`}>
                            {cluster.name}
                        </Link>,
                        cluster.description,
                        cluster.creator && <User user={cluster.creator} />,
                        formatDateTime(cluster.created_at),
                    ]) ?? []
                }
                paginationProps={{
                    start: clustersInfo.data?.start,
                    count: clustersInfo.data?.count,
                    total: clustersInfo.data?.total,
                    afterPageChange: () => {
                        clustersInfo.refetch()
                    },
                }}
            />
            <Modal
                isOpen={isCreateClusterOpen}
                onClose={() => setIsCreateClusterOpen(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('create sth', [t('cluster')])}</ModalHeader>
                <ModalBody>
                    <ClusterForm onSubmit={handleCreateCluster} />
                </ModalBody>
            </Modal>
        </Card>
    )
}
