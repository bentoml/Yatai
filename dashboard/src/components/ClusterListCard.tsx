import React, { useCallback, useState } from 'react'
import Card from '@/components/Card'
import { createCluster } from '@/services/cluster'
import { usePage } from '@/hooks/usePage'
import { ICreateClusterSchema } from '@/schemas/cluster'
import ClusterForm from '@/components/ClusterForm'
import { formatTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import User from '@/components/User'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import Table from '@/components/Table'
import { Link } from 'react-router-dom'
import { useFetchClusters } from '@/hooks/useFetchClusters'
import { resourceIconMapping } from '@/consts'

export interface IClusterListCardProps {
    orgName: string
}

export default function ClusterListCard({ orgName }: IClusterListCardProps) {
    const [page, setPage] = usePage()
    const clustersInfo = useFetchClusters(orgName, page)
    const [isCreateClusterOpen, setIsCreateClusterOpen] = useState(false)
    const handleCreateCluster = useCallback(
        async (data: ICreateClusterSchema) => {
            await createCluster(orgName, data)
            await clustersInfo.refetch()
            setIsCreateClusterOpen(false)
        },
        [clustersInfo, orgName]
    )
    const [t] = useTranslation()

    return (
        <Card
            title={t('sth list', [t('cluster')])}
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
                        <Link key={cluster.uid} to={`/orgs/${orgName}/clusters/${cluster.name}`}>
                            {cluster.name}
                        </Link>,
                        cluster.description,
                        cluster.creator && <User user={cluster.creator} />,
                        formatTime(cluster.created_at),
                    ]) ?? []
                }
                paginationProps={{
                    start: clustersInfo.data?.start,
                    count: clustersInfo.data?.count,
                    total: clustersInfo.data?.total,
                    onPageChange: ({ nextPage }) => {
                        setPage({
                            ...page,
                            start: nextPage * page.count,
                        })
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
